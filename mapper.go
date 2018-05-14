package align

import (
	"encoding/json"
	//"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/juliangruber/go-intersect"
	"github.com/labstack/echo"
	"github.com/recursionpharma/go-csv-map"
)

// This is a dummy repository; in real life call this code would be replaced by an API querying the repository
// assumes tab-delimited file with header.
// Expects to find fields URL     Content         Paradata        Manual-Alignment	Learning-Area	Year
// Year and Learning-Area can contain multiple values; they are ";"-delimited
// Paradata contains JSON map of curriculum IDs to hits
// Manual Alignment contains JSON list of curriculum IDs aligned by expert in the repository

type repository_entry struct {
	Url             string
	Content         string
	Paradata        map[string]int
	ManualAlignment []string
	LearningArea    []string
	Year            []string
}

func read_repository(directory string) (map[string]repository_entry, error) {
	files, _ := filepath.Glob(directory + "/*.txt")
	if len(files) == 0 {
		log.Fatalln("No *.txt repository files found in input folder" + directory)
	}
	records := make([]map[string]string, 0)
	for _, filename := range files {
		buf, err := os.Open(filename)
		if err != nil {
			log.Printf("%s: ", filename)
			log.Fatalln(err)
		}
		defer buf.Close()
		reader := csvmap.NewReader(buf)
		reader.Reader.Comma = '\t'
		columns, err := reader.ReadHeader()
		if err != nil {
			log.Printf("%s: ", filename)
			log.Fatalln(err)
		}
		reader.Columns = columns
		records1, err := reader.ReadAll()
		if err != nil {
			log.Printf("%s: ", filename)
			log.Fatalln(err)
		}
		records = append(records, records1...)
	}
	return convert_to_repository(records), nil
}

// iterate through CSV, converting text to repository struct. Keyed on URL, will overwrite
// entries with the same URL
func convert_to_repository(csv []map[string]string) map[string]repository_entry {
	ret := make(map[string]repository_entry, 0)
	for _, record := range csv {
		paradata := make(map[string]int)
		alignments := make([]string, 0)
		json.Unmarshal([]byte(record["Paradata"]), &paradata)
		json.Unmarshal([]byte(record["manual-Alignment"]), &alignments)
		years := strings.Split(strings.Replace(record["Year"], "\"", "", -1), ";")
		sort.Slice(years, func(i, j int) bool { return years[i] > years[j] })
		areas := strings.Split(strings.Replace(record["Learning-Area"], "\"", "", -1), ";")
		sort.Slice(areas, func(i, j int) bool { return areas[i] > areas[j] })
		ret[record["URL"]] = repository_entry{
			Url:             record["URL"],
			Content:         record["Content"],
			Year:            years,
			LearningArea:    areas,
			Paradata:        paradata,
			ManualAlignment: alignments,
		}
	}
	return ret
}

// filter repository to match language area(s) and year level(s)
func filter_repository(repository map[string]repository_entry, learning_area []string, years []string) map[string]repository_entry {
	sort.Slice(years, func(i, j int) bool { return years[i] > years[j] })
	sort.Slice(learning_area, func(i, j int) bool { return years[i] > years[j] })
	ret := make(map[string]repository_entry)
	for k, v := range repository {
		overlap := intersect.Simple(years, v.Year)
		if len(overlap.([]interface{})) == 0 {
			continue
		}
		overlap = intersect.Simple(learning_area, v.LearningArea)
		if len(overlap.([]interface{})) == 0 {
			continue
		}
		ret[k] = v
	}
	return ret
}

var repository map[string]repository_entry

func Init() {
	var err error
	repository, err = read_repository("./repository/")
	if err != nil {
		log.Fatalln(err)
	}
	repository = filter_repository(repository, []string{"Science"}, []string{"8"})
	log.Printf("%+v\n", repository)
}

func Align(c echo.Context) error {
	/*
		var years, learning_area, text string
			learning_area = c.QueryParam("area")
			text = c.QueryParam("text")
			year = c.QueryParam("year")
			log.Printf("Area: %s\nYears: %s\nText: %s\n", learning_area, year, text)
			if learning_area == "" {
				err := fmt.Errorf("area parameter not supplied")
				c.String(http.StatusBadRequest, err.Error())
				return err
			}
			if text == "" {
				err := fmt.Errorf("text parameter not supplied")
				c.String(http.StatusBadRequest, err.Error())
				return err
			}
			if year == "" {
				year = "K,P,1,2,3,4,5,6,7,8,9,10,11,12"
			}
			classifier, err := train_curriculum(curriculum, learning_area, strings.Split(year, ","))
			if err != nil {
				c.String(http.StatusBadRequest, err.Error())
				return err
			}
			response := classify_text(classifier, curriculum_map, text)
	*/
	response := ""
	return c.JSON(http.StatusOK, response)
}
