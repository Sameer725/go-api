package data

import "time"

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"` // - directive omits the item from json
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`    // - omitempty omits the item if empty/falsy value
	Runtime   Runtime   `json:"runtime,omitempty"` // - string directive changes the field item to string
	Genres    []string  `json:",omitempty"`        // leaving 1st directive blank leave the filed title as it is
	Version   int32     `json:"version"`
}
