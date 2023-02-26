package main

import (
	"encoding/csv"
	"fmt"
	"github.com/gocolly/colly"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Person struct {
	id         string
	no         int
	name       string
	secondName string
	gender     string
	parentId   string
	desc       string
}

var (
	people []Person
	re     = regexp.MustCompile(`\[([^\[\]]*)\]`)
)

func main() {
	crawlDat()
}

func crawlDat() {
	c := colly.NewCollector()

	c.OnHTML("ul#treetest", func(e *colly.HTMLElement) {
		// extract information from the current HTML tag
		nodeText := e.ChildText("ul#treetest > li > a")
		id := e.ChildAttr("ul#treetest > li > a", "infostep")
		fmt.Println(id, nodeText)
		people = append(people, Person{id, 1, nodeText, "", "Trai", "", ""})

		// recursively process child elements
		e.ForEach("ul#treetest > li", func(_ int, child *colly.HTMLElement) {
			recursiveScrape(child, id)
		})
		writeDataToCSV()
	})

	c.Visit("https://nguyenductoc.com/vn/Xemphahe.aspx")
}

func writeDataToCSV() {
	file, err := os.Create("people.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.UseCRLF = true
	defer writer.Flush()

	// Write the header row
	err = writer.Write([]string{"id", "family_no", "name", "second_name", "gender", "parent_id", "desc", "birthday", "date_of_death"})
	if err != nil {
		panic(err)
	}

	// Write the data rows
	for _, person := range people {
		err := writer.Write([]string{person.id, strconv.Itoa(person.no), person.name, person.secondName, person.gender, person.parentId, person.desc, "", ""})
		if err != nil {
			panic(err)
		}
	}
}

func recursiveScrape(e *colly.HTMLElement, id string) {
	queryText := fmt.Sprintf("a[infostep='%s'] ~ ul > li > a", id)
	queryChild := fmt.Sprintf("a[infostep='%s'] ~ ul > li", id)
	infoStep := make(map[int]string)

	e.ForEach(queryText, func(index int, childNested *colly.HTMLElement) {
		gender := "Gái"
		desc := ""
		infoStep[index] = childNested.Attr("infostep")
		name := strings.Title(strings.ToLower(strings.TrimSpace(childNested.Text)))
		matchValues := re.FindAllString(name, -1)
		secondName := ""

		// get second name
		if len(matchValues) > 0 {
			secondName = strings.Trim(matchValues[0], "[")
			secondName = strings.Trim(secondName, "]")
			secondName = strings.TrimSpace(secondName)
			if strings.Index(name, "Con Bà") >= 0 {
				secondName = ""
			}
		}

		squareBracketIndex := strings.Index(name, "[")
		if squareBracketIndex >= 0 {
			name = name[:squareBracketIndex]
			if secondName == "Vô Danh" {
				name = fmt.Sprintf("%s %s", name, secondName)
				secondName = ""
			}
			if strings.HasPrefix(secondName, "Liệt Sỷ") {
				secondName = ""
				desc = "Liệt sĩ"
			}
			name = strings.TrimSpace(name)
		}

		if strings.HasPrefix(name, "Nguyễn Đức") {
			gender = "Trai"
		}

		people = append(people, Person{infoStep[index], index + 1, name, secondName, gender, id, desc})
	})

	// recursively process child elements
	e.ForEach(queryChild, func(index int, child *colly.HTMLElement) {
		recursiveScrape(child, infoStep[index])
	})
}
