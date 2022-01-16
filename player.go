package main

import (
	"context"
	"fmt"

	"github.com/zmb3/spotify/v2"
)

func Tonst() string {
	return "Good to go!"
}

/*
-
*/

func search(client *spotify.Client, term string) (*spotify.FullTrack, error) {
	result, err := client.Search(context.Background(), term, spotify.SearchTypeTrack)
	if err != nil {
		return nil, fmt.Errorf("Error from client.Search(): term=%v, err=%v", term, err)
	}

	fmt.Printf("client.Search(\"%v\") gave: %v\n", term, result.Tracks)

	client.QueueSong(context.Background(), result.Tracks.Tracks[0].ID)
	return &result.Tracks.Tracks[0], nil
}

// It's OK to pass a string for trackID
func queue(client *spotify.Client, trackID spotify.ID) error {
	err := client.QueueSong(context.Background(), trackID)
	return fmt.Errorf("Error from client.QueueSong(): trackID=%v, err=%v", trackID, err)
}
