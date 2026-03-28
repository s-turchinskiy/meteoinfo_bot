package service

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
	"gopkg.in/yaml.v3"
)

var (
	ErrNoCity          = errors.New("this city is not supported")
	errRowsTr1More8    = errors.New("rows count in tr1 > 8")
	errRowsTr3More8    = errors.New("rows count in tr3 > 8")
	errTableNotFinding = errors.New("table not found")
	errTBodyNil        = errors.New("tbody nil")
	errTr1Nil          = errors.New("tr1 nil")
	errTr2Nil          = errors.New("tr2 nil")
	errTr3Nil          = errors.New("tr3 nil")
	errTdDontChild     = errors.New("td dont child element")
	offsets            = map[string]int{
		"Понедельник": 1,
		"Вторник":     10,
		"Среда":       14,
		"Четверг":     11,
		"Пятница":     10,
		"Суббота":     10,
		"Воскресенье": 2,
	}
	spacesRunes = []rune("                    ")
)

type DataReceiver interface {
	Receive(url string) (data []byte, err error)
}
type MeteoinforuService struct {
	cities map[string]string
	client DataReceiver
}
type item struct {
	day string
	val string
}

func NewService(receiver DataReceiver, citiesPath string) (*MeteoinforuService, error) {
	cities, err := readYaml(citiesPath)
	if err != nil {
		return nil, fmt.Errorf("error unmarshal yaml with cities, error: %w", err)
	}

	srvc := &MeteoinforuService{
		cities: cities,
		client: receiver,
	}

	return srvc, nil
}

func (s *MeteoinforuService) GetWeather(city string) (result string, err error) {
	url, ok := s.cities[city]
	if !ok {
		return "", ErrNoCity
	}

	data, err := s.getDataFromSite(url)
	if err != nil {
		return "", err
	}

	return displayData(data), nil
}

func (s *MeteoinforuService) getDataFromSite(url string) (*[7]item, error) {
	body, err := s.client.Receive(url)
	if err != nil {
		return nil, fmt.Errorf("error get data from site %w", err)
	}

	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("error parse data from site %w", err)
	}

	table := findTable(doc)
	if table == nil {
		return nil, errTableNotFinding
	}
	data, err := dataFromTbody(table)
	if err != nil {
		return nil, fmt.Errorf("error parse table %w", err)
	}

	return data, nil
}

func displayData(data *[7]item) string {
	sb := strings.Builder{}
	for _, val := range data {
		sb.WriteString(val.day + string(spacesRunes[:offsets[val.day]]) + val.val + "\n")
	}
	return sb.String()
}

func findTable(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "table" && n.Parent.Data == "div" {
		for _, attr := range n.Parent.Attr {
			if attr.Key == "id" && attr.Val == "div_print_0" {
				return n
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		res := findTable(c)
		if res != nil {
			return res
		}
	}

	return nil
}

func dataFromTbody(table *html.Node) (*[7]item, error) {
	result := &[7]item{}
	for i := range result {
		result[i] = item{}
	}

	tbody := table.FirstChild
	if tbody == nil {
		return nil, errTBodyNil
	}
	tr1 := tbody.FirstChild
	if tr1 == nil {
		return nil, errTr1Nil
	}

	i := 0
	for c := tr1.FirstChild; c != nil; c = c.NextSibling {
		i++

		if i == 1 {
			continue
		}

		if i > 8 {
			return nil, errRowsTr1More8
		}

		if c.FirstChild == nil {
			return nil, errTdDontChild
		}
		result[i-2].day = c.FirstChild.Data
	}

	i = 0

	tr2 := tr1.NextSibling
	if tr2 == nil {
		return nil, errTr2Nil
	}

	tr3 := tr2.NextSibling
	if tr3 == nil {
		return nil, errTr3Nil
	}

	for c := tr3.FirstChild; c != nil; c = c.NextSibling {
		i++
		if i == 1 {
			continue
		}

		if i > 8 {
			return nil, errRowsTr3More8
		}

		result[i-2].val = c.FirstChild.FirstChild.Data
	}

	return result, nil
}

func readYaml(filename string) (map[string]string, error) {
	_, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}

	yamlFile, err := os.ReadFile(filepath.Clean(filename))
	if err != nil {
		return nil, err
	}

	var data map[string]string
	err = yaml.Unmarshal(yamlFile, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
