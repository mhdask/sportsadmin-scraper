package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog/log"
)

var (
	URL        = "http://stats.sportsadmin.dk/schedule.aspx?tournamentID=1783"
	timeformat = "02-01-2006 15:04"
)

type Week struct {
	Round   int
	Matches []MatchRow
}

type MatchRow struct {
	Date     time.Time
	HomeTeam Team
	AwayTeam Team
	Result   string
}

type Team struct {
	Name   string
	Logo   string
	Winner bool
}

func main() {
	schedule := getFullSchedule()
	rounds := weekSplitter(schedule)
	for _, v := range rounds {
		v.checkWinner()
	}

	fmt.Println(rounds[1].Matches)

}

func getFullSchedule() []MatchRow {
	var headings, row []string
	var rows [][]string
	var matches []MatchRow

	res, err := http.Get(URL)
	if err != nil {
		log.Error().Err(err).Msg("could not contact sportsadmin")
		return nil
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Error().Err(err).Msg("could not get document from response")
		return nil
	}

	doc.Find("table").Each(func(index int, tablehtml *goquery.Selection) {
		tablehtml.Find("tr").Each(func(indextr int, rowhtml *goquery.Selection) {
			rowhtml.Find("th").Each(func(indexth int, tableheading *goquery.Selection) {
				headings = append(headings, tableheading.Text())
			})
			rowhtml.Find("td").Each(func(indexth int, tablecell *goquery.Selection) {
				row = append(row, tablecell.Text())
			})
			rows = append(rows, row)
			row = nil
		})
	})

	for i := 1; i < len(rows); i++ {
		date, err := time.Parse(timeformat, rows[i][0]+" "+rows[i][1])
		if err != nil {
			log.Error().Err(err).Msg("could not parse time")
			continue
		}

		hometeam := Team{rows[i][3], "", false}
		awayteam := Team{rows[i][4], "", false}
		result := rows[i][5]
		match := MatchRow{
			Date:     date,
			HomeTeam: hometeam,
			AwayTeam: awayteam,
			Result:   result,
		}

		matches = append(matches, match)

	}

	return matches
}

func (w *Week) checkWinner() {
	for i, r := range w.Matches {

		if r.Result == "" {
			return
		}

		resultvalues := strings.Split(r.Result, " - ")
		homescore, err := strconv.Atoi(resultvalues[0])
		if err != nil {
			log.Error().Err(err).Msg("could not convert hometeam score to integer")
		}

		awayscore, err := strconv.Atoi(resultvalues[1])
		if err != nil {
			log.Error().Err(err).Msg("could not convert awayteam score to integers")
		}

		if homescore > awayscore {
			w.Matches[i].HomeTeam.Winner = true

		} else if homescore < awayscore {
			w.Matches[i].AwayTeam.Winner = true

		}
	}
}

func weekSplitter(matches []MatchRow) []Week {
	var roundmatches []MatchRow
	var weeks []Week
	starttime := matches[0].Date.Add(-time.Duration(matches[0].Date.Hour()) * time.Hour)
	endtime := starttime.Add(8 * 24 * time.Hour)

	var round = 0
	for _, v := range matches {
		if v.Date.After(starttime) && v.Date.Before(endtime) || v.Date.Day() == endtime.Day() || v.Date.Day() == starttime.Day() {
			roundmatches = append(roundmatches, v)
		} else {
			weeks = append(weeks, Week{round, roundmatches})
			roundmatches = nil
			roundmatches = append(roundmatches, v)
			round += 1
			starttime = starttime.Add(7 * 24 * time.Hour)
			endtime = endtime.Add(7 * 24 * time.Hour)
		}
	}

	weeks = append(weeks, Week{round, roundmatches})
	return weeks
}
