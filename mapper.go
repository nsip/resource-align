package align

import (
	"encoding/json"
	"fmt"
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
	// log.Printf("%+v\n", filter_repository(repository, []string{"Science"}, []string{"8"}))
}

type alignment struct {
	Expert        float64
	Usage         float64
	TextBased     float64
	WeightedTotal float64
}

func extract_alignments(item repository_entry, alignments map[string]*alignment) map[string]*alignment {
	for _, statement := range item.ManualAlignment {
		if _, ok := alignments[statement]; !ok {
			alignments[statement] = &alignment{Expert: 0, Usage: 0, TextBased: 0, WeightedTotal: 0}
		}
		alignments[statement].Expert = alignments[statement].Expert + 1
	}
	for statement, value := range item.Paradata {
		if _, ok := alignments[statement]; !ok {
			alignments[statement] = &alignment{Expert: 0, Usage: 0, TextBased: 0, WeightedTotal: 0}
		}
		alignments[statement].Usage = alignments[statement].Usage + float64(value)
	}
	// TODO: call classifier on resource text
	return alignments
}

func normalise_alignments(alignments map[string]*alignment) map[string]*alignment {
	max := 0.0
	for _, v := range alignments {
		if max > v.Expert {
			max = v.Expert
		}
	}
	if max > 0 {
		for k, _ := range alignments {
			alignments[k].Expert = alignments[k].Expert / max
		}
	}
	max = 0.0
	for _, v := range alignments {
		if max > v.Usage {
			max = v.Usage
		}
	}
	if max > 0 {
		for k, _ := range alignments {
			alignments[k].Usage = alignments[k].Usage / max
		}
	}
	// TODO: normalise classifier results: (i - min) / (max - min)
	for k, _ := range alignments {
		// TODO: introduce weights
		alignments[k].WeightedTotal = alignments[k].Expert + alignments[k].Usage + alignments[k].TextBased
	}
	return alignments
}

func Align(c echo.Context) error {
	var year, learning_area string
	learning_area = c.QueryParam("area")
	year = c.QueryParam("year")
	if learning_area == "" {
		err := fmt.Errorf("area parameter not supplied")
		c.String(http.StatusBadRequest, err.Error())
		return err
	}
	if year == "" {
		year = "K,P,1,2,3,4,5,6,7,8,9,10,11,12"
	}
	resources := filter_repository(repository,
		strings.Split(strings.Replace(learning_area, "\"", "", -1), ";"),
		strings.Split(strings.Replace(year, "\"", "", -1), ";"),
	)
	response := make(map[string]*alignment)
	// TODO: filter candidate content descriptions by year and area
	for _, item := range resources {
		response = extract_alignments(item, response)
		response = normalise_alignments(response)
	}
	return c.JSON(http.StatusOK, response)
}
