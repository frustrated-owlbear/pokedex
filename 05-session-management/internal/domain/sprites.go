package domain

import "fmt"

// ImageURL returns official artwork URL for a national dex id.
func ImageURL(id int) string {
	if id <= 0 {
		return ""
	}
	return fmt.Sprintf(
		"https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/official-artwork/%d.png",
		id,
	)
}
