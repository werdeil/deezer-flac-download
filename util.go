package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func getTitle(song resSongInfoData) string {
	if song.Version != "" {
		return strings.Join([]string{song.SngTitle, song.Version}, " ")
	} else {
		return song.SngTitle
	}
}

func getArtist(song resSongInfoData) string {
	artistNames := make([]string, 0)
	for _, artist := range song.Artists {
		artistNames = append(artistNames, artist.ArtName)
	}
	sort.Strings(artistNames)
	fullArtist := strings.Join(artistNames, ", ")
	return fullArtist
}

func getComposer(song resSongInfoData) string {
	if song.SngContributors.Composer != nil {
		composers := make([]string, 0)
		for _, name := range song.SngContributors.Composer {
			composers = append(composers, name)
		}
		return strings.Join(composers, ", ")
	} else {
		return ""
	}
}

// extractYear returns the 4-digit year from a date string if possible.
// It handles formats like "YYYY-MM-DD" or "YYYY" and returns an empty
// string when a year cannot be determined.
func extractYear(dateStr string) string {
	dateStr = strings.TrimSpace(dateStr)
	if len(dateStr) >= 4 {
		year := dateStr[:4]
		for _, r := range year {
			if r < '0' || r > '9' {
				return ""
			}
		}
		return year
	}
	return ""
}

// getAlbumGenres returns a comma-separated list of genre names from the album.
// Falls back to album.Label if no genre entries are present.
func getAlbumGenres(album resAlbum) string {
	if len(album.Genres.Data) > 0 {
		names := make([]string, 0, len(album.Genres.Data))
		for _, g := range album.Genres.Data {
			if strings.TrimSpace(g.Name) != "" {
				names = append(names, g.Name)
			}
		}
		if len(names) > 0 {
			return strings.Join(names, ", ")
		}
	}
	if strings.TrimSpace(album.Label) != "" {
		return album.Label
	}
	return ""
}

func SanitizePath(rawPath string) string {
	cleanPath := filepath.Clean(rawPath)
	replacements := map[string]string{
		"<":  "-",
		">":  "-",
		":":  "-",
		"\"": "-",
		"/":  "-",
		"\\": "-",
		"|":  "-",
		"?":  "-",
		"*":  "-",
	}
	for old, new := range replacements {
		cleanPath = strings.ReplaceAll(cleanPath, old, new)
	}
	return cleanPath
}

func getSongPath(song resSongInfoData, album resAlbum, config configuration, format string) string {
	trackNum, err := strconv.Atoi(song.TrackNumber)
	cleanArtist := SanitizePath(album.Artist.Name)
	cleanAlbumTitle := SanitizePath(song.AlbTitle)
	cleanSongTitle := SanitizePath(song.SngTitle)
	if err != nil {
		panic(err)
	}
	ext := "flac"
	if strings.HasPrefix(strings.ToUpper(format), "MP3") {
		ext = "mp3"
	}
	return fmt.Sprintf("%s/%s/%s - %s/%02d - %s.%s", config.DestDir,
		cleanArtist, cleanArtist, cleanAlbumTitle, trackNum, cleanSongTitle, ext)
}
