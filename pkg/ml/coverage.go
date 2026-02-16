package ml

import (
	"fmt"
	"io"
	"strings"
)

// regionRow holds a single row from the region distribution query.
type regionRow struct {
	group   string
	n       int
	domains int
}

// PrintCoverage analyzes seed coverage by region and domain, printing
// a report with bar chart visualization and gap recommendations.
func PrintCoverage(db *DB, w io.Writer) error {
	rows, err := db.QueryRows("SELECT count(*) AS total FROM seeds")
	if err != nil {
		return fmt.Errorf("count seeds: %w (run: core ml import-all first)", err)
	}
	if len(rows) == 0 {
		return fmt.Errorf("no seeds table found (run: core ml import-all first)")
	}
	total := toInt(rows[0]["total"])

	fmt.Fprintln(w, "LEM Seed Coverage Analysis")
	fmt.Fprintln(w, "==================================================")
	fmt.Fprintf(w, "\nTotal seeds: %d\n", total)

	// Region distribution.
	regionRows, err := queryRegionDistribution(db)
	if err != nil {
		return fmt.Errorf("query regions: %w", err)
	}

	fmt.Fprintln(w, "\nRegion distribution (underrepresented first):")
	avg := float64(total) / float64(len(regionRows))
	for _, r := range regionRows {
		barLen := int(float64(r.n) / avg * 10)
		if barLen > 40 {
			barLen = 40
		}
		bar := strings.Repeat("#", barLen)
		gap := ""
		if float64(r.n) < avg*0.5 {
			gap = "  <- UNDERREPRESENTED"
		}
		fmt.Fprintf(w, "  %-22s %6d  (%4d domains)  %s%s\n", r.group, r.n, r.domains, bar, gap)
	}

	// Top 10 domains.
	fmt.Fprintln(w, "\nTop 10 domains (most seeds):")
	topRows, err := db.QueryRows(`
		SELECT domain, count(*) AS n FROM seeds
		WHERE domain != '' GROUP BY domain ORDER BY n DESC LIMIT 10
	`)
	if err == nil {
		for _, row := range topRows {
			domain := strVal(row, "domain")
			n := toInt(row["n"])
			fmt.Fprintf(w, "  %-40s %5d\n", domain, n)
		}
	}

	// Bottom 10 domains.
	fmt.Fprintln(w, "\nBottom 10 domains (fewest seeds, min 5):")
	bottomRows, err := db.QueryRows(`
		SELECT domain, count(*) AS n FROM seeds
		WHERE domain != '' GROUP BY domain HAVING count(*) >= 5 ORDER BY n ASC LIMIT 10
	`)
	if err == nil {
		for _, row := range bottomRows {
			domain := strVal(row, "domain")
			n := toInt(row["n"])
			fmt.Fprintf(w, "  %-40s %5d\n", domain, n)
		}
	}

	fmt.Fprintln(w, "\nSuggested expansion areas:")
	fmt.Fprintln(w, "  - Japanese, Korean, Thai, Vietnamese (no seeds found)")
	fmt.Fprintln(w, "  - Hindi/Urdu, Bengali, Tamil (South Asian)")
	fmt.Fprintln(w, "  - Swahili, Yoruba, Amharic (Sub-Saharan Africa)")
	fmt.Fprintln(w, "  - Indigenous languages (Quechua, Nahuatl, Aymara)")

	return nil
}

// queryRegionDistribution returns seed counts grouped by normalized language
// region, ordered ascending (underrepresented first).
func queryRegionDistribution(db *DB) ([]regionRow, error) {
	rows, err := db.QueryRows(`
		SELECT
			CASE
				WHEN region LIKE '%cn%' THEN 'cn (Chinese)'
				WHEN region LIKE '%en-%' OR region LIKE '%en_para%' OR region LIKE '%para%' THEN 'en (English)'
				WHEN region LIKE '%ru%' THEN 'ru (Russian)'
				WHEN region LIKE '%de%' AND region NOT LIKE '%deten%' THEN 'de (German)'
				WHEN region LIKE '%es%' THEN 'es (Spanish)'
				WHEN region LIKE '%fr%' THEN 'fr (French)'
				WHEN region LIKE '%latam%' THEN 'latam (LatAm)'
				WHEN region LIKE '%africa%' THEN 'africa'
				WHEN region LIKE '%eu%' THEN 'eu (European)'
				WHEN region LIKE '%me%' AND region NOT LIKE '%premium%' THEN 'me (MidEast)'
				WHEN region LIKE '%multi%' THEN 'multilingual'
				WHEN region LIKE '%weak%' THEN 'weak-langs'
				ELSE 'other'
			END AS lang_group,
			count(*) AS n,
			count(DISTINCT domain) AS domains
		FROM seeds GROUP BY lang_group ORDER BY n ASC
	`)
	if err != nil {
		return nil, err
	}

	result := make([]regionRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, regionRow{
			group:   strVal(row, "lang_group"),
			n:       toInt(row["n"]),
			domains: toInt(row["domains"]),
		})
	}
	return result, nil
}
