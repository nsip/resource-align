package align

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	"github.com/juliangruber/go-intersect"
	"github.com/labstack/echo"
	"github.com/nsip/curriculum-align"
	"gopkg.in/fatih/set.v0"
)

// requires that github.com/nsip/curriculum-align be running as a webservice, /curricalign, on port :1576

// This is a dummy repository; in real life call this code would be replaced by an API querying the repository
// Paradata contains JSON map of curriculum IDs to hits
// Manual Alignment contains JSON list of curriculum IDs aligned by expert in the repository

type repository_entry struct {
	Url             string `json:"URL"`
	Content         string
	Paradata        map[string]int
	ManualAlignment []string `json:"Manual-Alignment"`
	LearningArea    []string `json:"Learning-Area"`
	Year            []string
}

func read_repository(directory string) (map[string]repository_entry, error) {
	files, _ := filepath.Glob(directory + "/*.json")
	if len(files) == 0 {
		log.Fatalln("No *.json curriculum files found in input folder" + directory)
	}
	ret := make(map[string]repository_entry, 0)
	for _, filename := range files {
		buf, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Printf("%s: ", filename)
			log.Fatalln(err)
		}
		var records1 []repository_entry
		json.Unmarshal(buf, &records1)
		for _, record := range records1 {
			ret[record.Url] = record
		}
	}
	//log.Printf("%+v", ret)
	return ret, nil
}

// filter repository to match language area(s) and year level(s)
func filter_repository(repository map[string]repository_entry, learning_area []string, years []string) map[string]repository_entry {
	sort.Slice(years, func(i, j int) bool { return years[i] > years[j] })
	sort.Slice(learning_area, func(i, j int) bool { return learning_area[i] > learning_area[j] })
	ret := make(map[string]repository_entry)
	for k, v := range repository {
		if len(years) > 0 {
			overlap := intersect.Simple(years, v.Year)
			if len(overlap.([]interface{})) == 0 {
				continue
			}
		}
		if len(learning_area) > 0 {
			overlap := intersect.Simple(learning_area, v.LearningArea)
			if len(overlap.([]interface{})) == 0 {
				continue
			}
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
	Url           string
	Statement     string
	Expert        float64
	Usage         float64
	TextBased     float64
	WeightedTotal float64
}

func get_curric_alignments(learning_area string, year string, text string) ([]align.AlignmentType, error) {
	resp, err := http.Get("http://localhost:1576/curricalign?area=" + learning_area + "&year=" + year + "&text=" + url.QueryEscape(text))
	log.Println("http://localhost:1576/curricalign?area=" + learning_area + "&year=" + year + "&text=" + url.QueryEscape(text))
	matches := make([]align.AlignmentType, 0)
	if err != nil {
		log.Println("Quit get_curric_alignments (1)")
		return matches, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Println("Quit get_curric_alignments (2)")
		return matches, err
	}
	json.Unmarshal([]byte(body), &matches)
	//log.Printf("%+v\n", matches)
	return matches, nil
}

func extract_alignments(item repository_entry, learning_area string, year string, filter *set.Set) map[string]*alignment {
	alignments := make(map[string]*alignment)
	// get filter: curriculum items specific to the given learning_area and year
	for _, statement := range item.ManualAlignment {
		if !filter.Has(statement) {
			continue
		}
		key := statement + ":" + item.Url
		if _, ok := alignments[key]; !ok {
			alignments[key] = &alignment{Expert: 0, Usage: 0, TextBased: 0, WeightedTotal: 0, Url: item.Url, Statement: statement}
		}
		alignments[key].Expert = alignments[key].Expert + 1
	}
	out, _ := json.MarshalIndent(alignments, "", "  ")
	for statement, value := range item.Paradata {
		if !filter.Has(statement) {
			continue
		}
		key := statement + ":" + item.Url
		if _, ok := alignments[key]; !ok {
			alignments[key] = &alignment{Expert: 0, Usage: 0, TextBased: 0, WeightedTotal: 0, Url: item.Url, Statement: statement}
		}
		alignments[key].Usage = alignments[key].Usage + float64(value)
	}
	out, _ = json.MarshalIndent(alignments, "", "  ")
	matches, err := get_curric_alignments(learning_area, year, item.Content)
	// if err, we ignore
	if err == nil {
		i := 0
		for _, match := range matches {
			if !filter.Has(match.Item) {
				continue
			}
			i += 1
			key := match.Item + ":" + item.Url
			if _, ok := alignments[key]; !ok {
				alignments[key] = &alignment{Expert: 0, Usage: 0, TextBased: 0, WeightedTotal: 0, Url: item.Url, Statement: match.Item}
			}
			alignments[key].TextBased = match.Score
		}
	} else {
		log.Println("FAIL: http://localhost:1576/curricalign?area=" + learning_area + "&year=" + year + "&text=" + url.QueryEscape(item.Content))
		log.Println(err)
	}
	out, _ = json.MarshalIndent(alignments, "", "  ")
	return alignments
}

func normalise_alignments(alignments map[string]*alignment) map[string]*alignment {
	max := 0.0
	for _, v := range alignments {
		if max < v.Expert {
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
		if max < v.Usage {
			max = v.Usage
		}
	}
	if max > 0 {
		for k, _ := range alignments {
			alignments[k].Usage = alignments[k].Usage / max
		}
	}
	// normalise classifier results: (i - min) / (max - min)
	// classifier results are negative
	max = -999999999.0
	min := 0.0
	for _, v := range alignments {
		if max < v.TextBased {
			max = v.TextBased
		}
		if min > v.TextBased {
			min = v.TextBased
		}
	}
	if min < 0 {
		for k, _ := range alignments {
			alignments[k].TextBased = (alignments[k].TextBased - min) / (max - min)

		}
	}
	for k, _ := range alignments {
		// TODO: introduce weights
		alignments[k].WeightedTotal = alignments[k].Expert + alignments[k].Usage + alignments[k].TextBased
	}
	return alignments
}

func alignments_to_sorted_array(alignments []map[string]*alignment) []alignment {
	ret := make([]alignment, 0)
	for _, a := range alignments {
		for _, v := range a {
			ret = append(ret, *v)
		}
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].WeightedTotal > ret[j].WeightedTotal })
	return ret
}

func rank_resources(alignments []alignment) []alignment {
	items := set.New()
	ret := make([]alignment, 0)
	for _, a := range alignments {
		if !items.Has(a.Url) {
			ret = append(ret, a)
			items.Add(a.Url)
		}
	}
	return ret
}

func param2slice(q string) []string {
	ret1 := strings.Split(strings.Replace(q, "\"", "", -1), ",")
	ret := make([]string, 0)
	for _, a := range ret1 {
		if len(a) > 0 {
			ret = append(ret, a)
		}
	}
	return ret
}

func Align(c echo.Context) error {
	var year, learning_area, item string
	learning_area = c.QueryParam("area")
	year = c.QueryParam("year")
	item = c.QueryParam("item")
	if year == "" {
		year = "K,P,1,2,3,4,5,6,7,8,9,10,11,12"
	}
	resources := filter_repository(repository, param2slice(learning_area), param2slice(year))
	items_arr := strings.Split(strings.Replace(item, "\"", "", -1), ",")
	items_set := set.New()
	for _, a := range items_arr {
		if len(a) > 0 {
			items_set.Add(a)
		}
	}
	// filter candidate content descriptions by year and area
	matches, _ := get_curric_alignments(learning_area, year, "a a a a a a a a")
	curric_filter := set.New()
	//log.Printf("%+v\n", matches)
	for _, item := range matches {
		if items_set.IsEmpty() || items_set.Has(item.Item) {
			curric_filter.Add(item.Item)
		}
	}
	log.Printf("Filtering against Year/Learning Area: only report on: %+v\n", curric_filter)

	resource_arr := make([]map[string]*alignment, 0)
	for _, item := range resources {
		log.Println("ITEM::")
		log.Printf("%+v\n", item)
		response := normalise_alignments(extract_alignments(item, learning_area, year, curric_filter))
		log.Println("NORMALISE::")
		out, _ := json.MarshalIndent(alignments_to_sorted_array([]map[string]*alignment{response}), "", "  ")
		log.Println(string(out))
		resource_arr = append(resource_arr, response)
	}
	response1 := alignments_to_sorted_array(resource_arr)
	out, _ := json.MarshalIndent(response1, "", "  ")
	log.Println(string(out))
	response1 = rank_resources(response1)
	return c.JSONPretty(http.StatusOK, response1, "  ")
}
