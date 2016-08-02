package main

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	//"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	fmt.Println("Search for=" + os.Args[1])
	resp, err := http.Get("http://search2.ucl.ac.uk/s/search.json?collection=website-meta&profile=_directory&tab=directory&num_ranks=1000&query=" + os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	type ucl_MetaData struct {
		Title      string `json:"a"`
		Name       string `json:"B"`
		Surname    string `json:"s"`
		Email      string `json:"E"`
		Tel        string `json:"M"`
		Extension  string `json:"W"`
		Department string `json:"7"`
		UPI        string `json:"U"`
		Type       string `json:"g"`
	}
	type Result struct {
		Title string `json:"title"`
		//Summary    string       `json:"summary"`
		//DisplayUrl string       `json:"displayUrl"`
		MetaData ucl_MetaData `json:"metaData"`
	}
	type ResultPacket struct {
		Results []Result `json:"results"`
	}
	type Response struct {
		MyResultPacket ResultPacket `json:"resultPacket"`
	}
	type Answer struct {
		MyResponse Response `json:"response"`
	}
	//respArray, _ := ioutil.ReadAll(resp.Body)
	//fmt.Printf("%s\n", string(respArray))

	dec := json.NewDecoder(resp.Body)
	var m Answer
	parse_err := dec.Decode(&m)
	if parse_err != nil {
		log.Fatal(parse_err)
	}

	ucl_people := m.MyResponse.MyResultPacket.Results
	var ucl_person ucl_MetaData

	//fmt.Printf("UCL response: %+v", ucl_people)

	if ucl_people == nil {
		log.Fatal("Not found in UCL directory")
	}

	for p := range ucl_people {
		fmt.Printf("UCL directory: %s, %s\n", ucl_people[p].MetaData.Surname, ucl_people[p].MetaData.Name)
	}
	if len(ucl_people) > 1 {
		log.Fatal("More than one person found....")
	} else {
		ucl_person = ucl_people[0].MetaData
	}

	//Create a list of organisation IDs that refer to UCL
	//Store sting of the University College London ID
	//uclOrgs := []string{"http://gtr.rcuk.ac.uk:80/gtr/api/organisations/3A5E126D-C175-4730-9B7B-E6D8CF447F83"}
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://gtr.rcuk.ac.uk/gtr/api/organisations?p=1&s=100&q=UCL&f=org.n", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Accept", "application/json")
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	type gtr_Link struct {
		Ref string `json:"href"`
		Rel string `json:"rel"`
	}
	type gtr_Identifier struct {
		Value string `json:"value"`
		Type  string `json:"type"`
	}
	type gtr_Links struct {
		Link []gtr_Link `json:"link"`
	}
	type gtr_Identifiers struct {
		Identifier []gtr_Identifier `json:"identifier"`
	}
	type gtr_Organisation struct {
		Name  string    `json:"name"`
		ID    string    `json:"id"`
		Links gtr_Links `json:"links"`
	}
	type gtr_OrgResponse struct {
		Page          int                `json:"page"`
		Size          int                `json:"size"`
		TotalPages    int                `json:"totalPages"`
		TotalSize     int                `json:"totalSize"`
		Organisations []gtr_Organisation `json:"organisation"`
	}
	type gtr_Person struct {
		ID         string    `json:"id"`
		FirstName  string    `json:"firstName"`
		OtherNames string    `json:"otherNames"`
		Surname    string    `json:"surname"`
		Links      gtr_Links `json:"links"`
	}
	type gtr_PersonResponse struct {
		Page       int          `json:"page"`
		Size       int          `json:"size"`
		TotalPages int          `json:"totalPages"`
		TotalSize  int          `json:"totalSize"`
		Persons    []gtr_Person `json:"person"`
	}
	type gtr_Project struct {
		ID          string          `json:"id"`
		Links       gtr_Links       `json:"links"`
		Identifiers gtr_Identifiers `json:"identifiers"`
		Title       string          `json:"title"`
		Abstract    string          `json:"abstractText"`
		Status      string          `json:"status"`
		Category    string          `json:"grantCategory"`
		Impact      string          `json:"potentialImpact"`
	}
	type gtr_ProjectResponse struct {
		Page       int           `json:"page"`
		Size       int           `json:"size"`
		TotalPages int           `json:"totalPages"`
		TotalSize  int           `json:"totalSize"`
		Projects   []gtr_Project `json:"project"`
	}

	//{"page":1,"size":30,"totalPages":1,"totalSize":30,"organisation":[
	//output, err := ioutil.ReadAll(res.Body)
	//res.Body.Close()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Printf("%s\n", output)
	dec = json.NewDecoder(res.Body)
	var r gtr_OrgResponse
	parse_err = dec.Decode(&r)
	if parse_err != nil {
		log.Fatal(parse_err)
	}
	orgs := r.Organisations
	//Map of ID to ?UCL
	isUCL := make(map[string]bool)
	isUCL["http://gtr.rcuk.ac.uk:80/gtr/api/organisations/3A5E126D-C175-4730-9B7B-E6D8CF447F83"] = true
	for o := range orgs {
		//fmt.Printf("\n%s %s", orgs[o].Name, orgs[o].ID)
		isUCL["http://gtr.rcuk.ac.uk:80/gtr/api/organisations/"+orgs[o].ID] = true
	}

	//Query all persons and search for employment?
	//http://gtr.rcuk.ac.uk/gtr/api/persons
	req, err = http.NewRequest("GET", "http://gtr.rcuk.ac.uk/gtr/api/persons?p=1&s=100&q="+ucl_person.Surname+"&f=per.sn", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Accept", "application/json")
	res, err = client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	dec = json.NewDecoder(res.Body)
	var pr gtr_PersonResponse
	parse_err = dec.Decode(&pr)
	if parse_err != nil {
		log.Fatal(parse_err)
	}
	people := pr.Persons
	uclPeople := make([]gtr_Person, 0)
	//Find the EMPLOYED link and search for href in isUCL
	for p := range people {
		links := people[p].Links.Link
		for l := range links {
			if links[l].Rel == "EMPLOYED" {
				if isUCL[links[l].Ref] {
					uclPeople = append(uclPeople, people[p])
					//The links also could contain COI_PER or PI_PER
				}
			}
		}
	}
	for p := range uclPeople {
		fmt.Printf("GtR: %s, %s\n", uclPeople[p].Surname, uclPeople[p].FirstName)
		links := uclPeople[p].Links.Link
		for l := range links {
			if strings.Contains(links[l].Rel, "PER") {
				//links[l].Ref is a project, I hope
				req, err = http.NewRequest("GET", links[l].Ref, nil)
				if err != nil {
					log.Fatal(err)
				}
				req.Header.Add("Accept", "application/json")
				res, err = client.Do(req)
				if err != nil {
					log.Fatal(err)
				}
				//output, err := ioutil.ReadAll(res.Body)
				//res.Body.Close()
				//if err != nil {
				//		log.Fatal(err)
				//}
				//fmt.Printf("%s\n", output)

				dec = json.NewDecoder(res.Body)
				var project gtr_Project
				parse_err = dec.Decode(&project)
				if parse_err != nil {
					log.Fatal(parse_err)
				}
				bt := color.New(color.Bold, color.Underline)
				bt.Printf("\n\nTitle: ")
				fmt.Println(project.Title)
				bt.Printf("Abstract: ")
				fmt.Println(project.Abstract)
				bt.Printf("Status: ")
				fmt.Println(project.Status)
				bt.Printf("Category: ")
				fmt.Println(project.Category)
				bt.Printf("Identifiers: ")
				fmt.Printf("%+v\n", project.Identifiers)
			}
		}
	}
}
