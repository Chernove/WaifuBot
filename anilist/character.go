package anilist

import (
	"context"

	"github.com/machinebox/graphql"
)

const graphURL string = "https://graphql.anilist.co"

// CharSearchStruct handles data from CharByName queries
type CharSearchStruct struct {
	Character CharacterStruct
}

// CharStruct handles data for RandomChar function
type CharStruct struct {
	Page struct {
		Characters []CharacterStruct
	}
}

// CharacterStruct represent character object
type CharacterStruct struct {
	SiteURL string `json:"siteUrl"`
	Image   struct {
		Large string `json:"large"`
	}
	Name struct {
		Full string `json:"full"`
	}
	Media struct {
		Nodes []struct {
			Title struct {
				Romaji string `json:"romaji"`
			}
		}
	}
	ID int64 `json:"id"`
}

// CharSearchInput is used to input the arguments you want to search
type CharSearchInput struct {
	Name string
	ID   int
}

// CharSearch makes a query to the Anilist API based on the name/ID you input
func CharSearch(input CharSearchInput) (response CharSearchStruct, err error) {
	// build query
	req := graphql.NewRequest(`
	query ($id: Int, $name: String) {
		Character(id: $id, search: $name, sort: SEARCH_MATCH) {
		  id
		  siteUrl
		  name {
			full
		  }
		  image {
			large
		  }
		  media(perPage: 1, sort: POPULARITY_DESC) {
			nodes {
			  title {
				romaji
			  }
			}
		  }
		}
	  }
	`)

	if input.ID != 0 {
		req.Var("id", input.ID)
	} else {
		req.Var("name", input.Name)
	}
	err = graphql.NewClient(graphURL).Run(context.Background(), req, &response)
	return
}

// CharSearchByPopularity outputs the character you want based on their number on the page list
func CharSearchByPopularity(id uint64, notIn []int64) (response CharStruct, err error) {
	// Create request
	req := graphql.NewRequest(`
	query ($pageNumber: Int, $not_in: [Int]) {
		Page(perPage: 1, page: $pageNumber) {
		  characters(sort: FAVOURITES_DESC, id_not_in: $not_in) {
			id
			siteUrl
			image {
			  large
			}
			name {
			  full
			}
			media(perPage: 1, sort: POPULARITY_DESC) {
			  nodes {
				title {
				  romaji
				}
			  }
			}
		  }
		}
	  }	  
	`)

	req.Var("pageNumber", id)
	if len(notIn) > 0 {
		req.Var("not_in", notIn)
	}

	// Make request
	err = graphql.NewClient(graphURL).Run(context.Background(), req, &response)
	return
}
