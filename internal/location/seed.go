package location

import (
	"context"

	"yanxo/internal/repository"
)

// SeedData: canonical name -> list of user-typed variants (will be normalized for alias storage).
var seedData = []struct {
	Canonical string
	Aliases   []string
}{
	{"Toshkent", []string{"toshkent", "tashkent", "ташкент", "tosheknt"}},
	{"Samarqand", []string{"samarqand", "samarkand", "самарқанд", "samarqand"}},
	{"Xiva", []string{"xiva", "hiva", "хива", "khiva"}},
	{"Xovos", []string{"xovos", "hovos", "ховос", "khovos"}},
	{"Yangiyer", []string{"yangiyer", "yangier", "янгиер", "yangiyo‘r"}},
	{"Guliston", []string{"guliston", "гулистан"}},
	{"Sirdaryo", []string{"sirdaryo", "sirdarya", "сирдарё"}},
	{"Qo‘qon", []string{"qoqon", "qo'qon", "kokand", "қўқон", "quqon"}},
	{"Andijon", []string{"andijon", "andijan", "андижон"}},
	{"Farg‘ona", []string{"fargona", "farg'ona", "fergana", "фарғона"}},
	{"Namangan", []string{"namangan", "наманган"}},
	{"Jizzax", []string{"jizzax", "jizzakh", "жизах", "dzhizak"}},
	{"Buxoro", []string{"buxoro", "bukhara", "бухоро"}},
	{"Navoiy", []string{"navoiy", "navoi", "навоий"}},
	{"Nukus", []string{"nukus", "нукус"}},
	{"Termiz", []string{"termiz", "termez", "термиз"}},
}

// SeedLocations idempotently inserts canonical locations and their normalized aliases.
func SeedLocations(ctx context.Context, repo repository.LocationRepository) error {
	for _, row := range seedData {
		aliases := make([]string, 0, len(row.Aliases)+1)
		aliases = append(aliases, Normalize(row.Canonical))
		for _, a := range row.Aliases {
			n := Normalize(a)
			if n != "" {
				aliases = append(aliases, n)
			}
		}
		// Dedupe
		seen := make(map[string]bool)
		var unique []string
		for _, a := range aliases {
			if !seen[a] {
				seen[a] = true
				unique = append(unique, a)
			}
		}
		if err := repo.EnsureLocationWithAliases(ctx, row.Canonical, unique); err != nil {
			return err
		}
	}
	return nil
}
