package main

import (
	"bytes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/blowfish"

	"github.com/BurntSushi/toml"
	id3v2 "github.com/bogem/id3v2"
	"github.com/go-flac/flacpicture"
	"github.com/go-flac/flacvorbis"
	"github.com/go-flac/go-flac"
)

type configuration struct {
	Arl          string `toml:"arl"`
	LicenseToken string `toml:"license_token"`
	DestDir      string `toml:"dest_dir"`
	Iv           string `toml:"iv"`
	PreKey       string `toml:"pre_key"`
}

type resTrackAlbum struct {
	Id          int64  `json:"id"`
	Title       string `json:"title"`
	Cover       string `json:"cover"`
	CoverSmall  string `json:"cover_small"`
	CoverMedium string `json:"cover_medium"`
	CoverBig    string `json:"cover_big"`
	CoverXl     string `json:"cover_xl"`
	Md5Image    string `json:"md5_image"`
	Tracklist   string `json:"tracklist"`
	Type        string `json:"type"`
}

type resTrackArtist struct {
	Id            int64  `json:"id"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	PictureSmall  string `json:"picture_small"`
	PictureMedium string `json:"picture_medium"`
	PictureBig    string `json:"picture_big"`
	PictureXl     string `json:"picture_xl"`
	Tracklist     string `json:"tracklist"`
	Type          string `json:"type"`
}

type resTrack struct {
	Id                    int64          `json:"id"`
	Readable              bool           `json:"readable"`
	Title                 string         `json:"title"`
	Link                  string         `json:"link"`
	Duration              int            `json:"duration"`
	Rank                  int            `json:"rank"`
	ExplicitLyrics        bool           `json:"explicit_lyrics"`
	ExplicitContentLyrics int            `json:"explicit_content_lyrics"`
	ExplicitContentCover  int            `json:"explicit_content_cover"`
	Md5Image              string         `json:"md5_image"`
	TimeAdd               int64          `json:"time_add"`
	Album                 resTrackAlbum  `json:"album"`
	Artist                resTrackArtist `json:"artist"`
	Type                  string         `json:"type"`
}

type resTracks struct {
	Data  []resTrack `json:"data"`
	Total int        `json:"total"`
}

type resSongInfoArtist struct {
	ArtId             string      `json:"ART_ID"`
	RoleId            string      `json:"ROLE_ID"`
	ArtistsSongsOrder string      `json:"ARTISTS_SONGS_ORDER"`
	ArtName           string      `json:"ART_NAME"`
	ArtistIsDummy     bool        `json:"ARTIST_IS_DUMMY"`
	ArtPicture        string      `json:"ART_PICTURE"`
	Rank              string      `json:"RANK"`
	Locales           interface{} `json:"LOCALES"`
	Type              string      `json:"__TYPE__"`
}

type resSongInfoMedia struct {
	Type string `json:"TYPE"`
	Href string `json:"HREF"`
}

type resSongInfoRights struct {
	StreamAdsAvailable bool   `json:"STREAM_ADS_AVAILABLE"`
	StreamAds          string `json:"STREAM_ADS"`
	StreamSubAvailable bool   `json:"STREAM_SUB_AVAILABLE"`
	StreamSub          string `json:"STREAM_SUB"`
}

type resSongInfoContributors struct {
	MainArtist     []string `json:"main_artist"`
	Composer       []string `json:"composer"`
	Featuring      []string `json:"featuring"`
	Narrator       []string `json:"narrator"`
	MusicPublisher []string `json:"music publisher"`
}

type resSongInfoExplicitTrackContent struct {
	ExplicitLyricsStatus int `json:"EXPLICIT_LYRICS_STATUS"`
	ExplicitCoverStatus  int `json:"EXPLICIT_COVER_STATUS"`
}

type resSongInfoAvailableCountries struct {
	StreamAds     []string      `json:"STREAM_ADS"`
	StreamSubOnly []interface{} `json:"STREAM_SUB_ONLY"`
}

type resSongInfoData struct {
	SngId                string                          `json:"SNG_ID"`
	ProductTrackId       string                          `json:"PRODUCT_TRACK_ID"`
	UploadId             int                             `json:"UPLOAD_ID"`
	SngTitle             string                          `json:"SNG_TITLE"`
	ArtId                string                          `json:"ART_ID"`
	ProviderId           string                          `json:"PROVIDER_ID"`
	ArtName              string                          `json:"ART_NAME"`
	ArtistIsDummy        bool                            `json:"ARTIST_IS_DUMMY"`
	Artists              []resSongInfoArtist             `json:"ARTISTS"`
	AlbId                string                          `json:"ALB_ID"`
	AlbTitle             string                          `json:"ALB_TITLE"`
	Type                 int                             `json:"TYPE"`
	Md5Origin            string                          `json:"MD5_ORIGIN"`
	Video                bool                            `json:"VIDEO"`
	Duration             string                          `json:"DURATION"`
	AlbPicture           string                          `json:"ALB_PICTURE"`
	ArtPicture           string                          `json:"ART_PICTURE"`
	RankSng              string                          `json:"RANK_SNG"`
	FilesizeAac64        string                          `json:"FILESIZE_AAC_64"`
	FilesizeMp364        string                          `json:"FILESIZE_MP3_64"`
	FilesizeMp3128       string                          `json:"FILESIZE_MP3_128"`
	FilesizeMp3256       string                          `json:"FILESIZE_MP3_256"`
	FilesizeMp3320       string                          `json:"FILESIZE_MP3_320"`
	FilesizeFlac         string                          `json:"FILESIZE_FLAC"`
	Filesize             string                          `json:"FILESIZE"`
	Gain                 string                          `json:"GAIN"`
	MediaVersion         string                          `json:"MEDIA_VERSION"`
	DiskNumber           string                          `json:"DISK_NUMBER"`
	TrackNumber          string                          `json:"TRACK_NUMBER"`
	TrackToken           string                          `json:"TRACK_TOKEN"`
	TrackTokenExpire     int                             `json:"TRACK_TOKEN_EXPIRE"`
	Version              string                          `json:"VERSION"`
	Media                []resSongInfoMedia              `json:"MEDIA"`
	ExplicitLyrics       string                          `json:"EXPLICIT_LYRICS"`
	Rights               resSongInfoRights               `json:"RIGHTS"`
	Isrc                 string                          `json:"ISRC"`
	HierarchicalTitle    string                          `json:"HIERARCHICAL_TITLE"`
	SngContributors      resSongInfoContributors         `json:"SNG_CONTRIBUTORS"`
	LyricsId             int                             `json:"LYRICS_ID"`
	ExplicitTrackContent resSongInfoExplicitTrackContent `json:"EXPLICIT_TRACK_CONTENT"`
	Copyright            string                          `json:"COPYRIGHT"`
	PhysicalReleaseDate  string                          `json:"PHYSICAL_RELEASE_DATE"`
	SMod                 int                             `json:"S_MOD"`
	SPremium             int                             `json:"S_PREMIUM"`
	DateStartPremium     string                          `json:"DATE_START_PREMIUM"`
	DateStart            string                          `json:"DATE_START"`
	Status               int                             `json:"STATUS"`
	UserId               int                             `json:"USER_ID"`
	URLRewriting         string                          `json:"URL_REWRITING"`
	SngStatus            string                          `json:"SNG_STATUS"`
	AvailableCountries   resSongInfoAvailableCountries   `json:"AVAILABLE_COUNTRIES"`
	UpdateDate           string                          `json:"UPDATE_DATE"`
	Type0                string                          `json:"__TYPE__"`
	DigitalReleaseDate   string                          `json:"DIGITAL_RELEASE_DATE"`
}

type resSongInfoIsrcData struct {
	ArtName            string            `json:"ART_NAME"`
	ArtId              string            `json:"ART_ID"`
	AlbPicture         string            `json:"ALB_PICTURE"`
	AlbId              string            `json:"ALB_ID"`
	AlbTitle           string            `json:"ALB_TITLE"`
	Duration           string            `json:"DURATION"`
	DigitalReleaseDate string            `json:"DIGITAL_RELEASE_DATE"`
	Rights             resSongInfoRights `json:"RIGHTS"`
	LyricsId           int               `json:"LYRICS_ID"`
	Type               string            `json:"__TYPE__"`
}

type resSongInfoIsrc struct {
	Data  []resSongInfoIsrcData `json:"data"`
	Count int                   `json:"count"`
	Total int                   `json:"total"`
}

type resSongInfoRelatedAlbumsData struct {
	ArtName            string            `json:"ART_NAME"`
	ArtId              string            `json:"ART_ID"`
	AlbPicture         string            `json:"ALB_PICTURE"`
	AlbId              string            `json:"ALB_ID"`
	AlbTitle           string            `json:"ALB_TITLE"`
	Duration           string            `json:"DURATION"`
	DigitalReleaseDate string            `json:"DIGITAL_RELEASE_DATE"`
	Rights             resSongInfoRights `json:"RIGHTS"`
	LyricsId           int               `json:"LYRICS_ID"`
	Type               string            `json:"__TYPE__"`
}

type resSongInfoRelatedAlbums struct {
	Data  []resSongInfoRelatedAlbumsData `json:"data"`
	Count int                            `json:"count"`
	Total int                            `json:"total"`
}

type resSongInfo struct {
	Data          resSongInfoData          `json:"DATA"`
	Isrc          resSongInfoIsrc          `json:"ISRC"`
	RelatedAlbums resSongInfoRelatedAlbums `json:"RELATED_ALBUMS"`
}

type resSongUrl struct {
	Data []struct {
		Errors []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
		Media []struct {
			Cipher struct {
				Type string `json:"type"`
			} `json:"cipher"`
			Exp       int    `json:"exp"`
			Format    string `json:"format"`
			MediaType string `json:"media_type"`
			Nbf       int    `json:"nbf"`
			Sources   []struct {
				Provider string `json:"provider"`
				Url      string `json:"url"`
			} `json:"sources"`
		} `json:"media"`
	} `json:"data"`
}

// This struct does not have all the fields that exist in the JSON
// because we only care about SONGS at the moment
type resAlbumInfo struct {
	Songs struct {
		Data          []resSongInfoData `json:"data"`
		Count         int               `json:"count"`
		Total         int               `json:"total"`
		FilteredCount int               `json:"filtered_count"`
	} `json:"SONGS"`
}

type resAlbumGenres struct {
	Data []struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
		Type    string `json:"type"`
	} `json:"data"`
}

type resAlbumContributor struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Link          string `json:"link"`
	Share         string `json:"share"`
	Picture       string `json:"picture"`
	PictureSmall  string `json:"picture_small"`
	PictureMedium string `json:"picture_medium"`
	PictureBig    string `json:"picture_big"`
	PictureXl     string `json:"picture_xl"`
	Radio         bool   `json:"radio"`
	Tracklist     string `json:"tracklist"`
	Type          string `json:"type"`
	Role          string `json:"role"`
}

type resAlbumArtist struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	PictureSmall  string `json:"picture_small"`
	PictureMedium string `json:"picture_medium"`
	PictureBig    string `json:"picture_big"`
	PictureXl     string `json:"picture_xl"`
	Tracklist     string `json:"tracklist"`
	Type          string `json:"type"`
}

type resAlbumTracks struct {
	Data []struct {
		ID                    int    `json:"id"`
		Readable              bool   `json:"readable"`
		Title                 string `json:"title"`
		TitleShort            string `json:"title_short"`
		TitleVersion          string `json:"title_version"`
		Link                  string `json:"link"`
		Duration              int    `json:"duration"`
		Rank                  int    `json:"rank"`
		ExplicitLyrics        bool   `json:"explicit_lyrics"`
		ExplicitContentLyrics int    `json:"explicit_content_lyrics"`
		ExplicitContentCover  int    `json:"explicit_content_cover"`
		Preview               string `json:"preview"`
		Md5Image              string `json:"md5_image"`
		Artist                struct {
			ID        int    `json:"id"`
			Name      string `json:"name"`
			Tracklist string `json:"tracklist"`
			Type      string `json:"type"`
		} `json:"artist"`
		Album struct {
			ID          int    `json:"id"`
			Title       string `json:"title"`
			Cover       string `json:"cover"`
			CoverSmall  string `json:"cover_small"`
			CoverMedium string `json:"cover_medium"`
			CoverBig    string `json:"cover_big"`
			CoverXl     string `json:"cover_xl"`
			Md5Image    string `json:"md5_image"`
			Tracklist   string `json:"tracklist"`
			Type        string `json:"type"`
		} `json:"album"`
		Type string `json:"type"`
	} `json:"data"`
}

type resAlbum struct {
	ID                    int                   `json:"id"`
	Title                 string                `json:"title"`
	Upc                   string                `json:"upc"`
	Link                  string                `json:"link"`
	Share                 string                `json:"share"`
	Cover                 string                `json:"cover"`
	CoverSmall            string                `json:"cover_small"`
	CoverMedium           string                `json:"cover_medium"`
	CoverBig              string                `json:"cover_big"`
	CoverXl               string                `json:"cover_xl"`
	Md5Image              string                `json:"md5_image"`
	GenreID               int                   `json:"genre_id"`
	Genres                resAlbumGenres        `json:"genres"`
	Label                 string                `json:"label"`
	NbTracks              int                   `json:"nb_tracks"`
	Duration              int                   `json:"duration"`
	Fans                  int                   `json:"fans"`
	ReleaseDate           string                `json:"release_date"`
	RecordType            string                `json:"record_type"`
	Available             bool                  `json:"available"`
	Tracklist             string                `json:"tracklist"`
	ExplicitLyrics        bool                  `json:"explicit_lyrics"`
	ExplicitContentLyrics int                   `json:"explicit_content_lyrics"`
	ExplicitContentCover  int                   `json:"explicit_content_cover"`
	Contributors          []resAlbumContributor `json:"contributors"`
	Artist                resAlbumArtist        `json:"artist"`
	Type                  string                `json:"type"`
	Tracks                resAlbumTracks        `json:"tracks"`
}

type resPlaylist struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Picture   string    `json:"picture"`
	PictureXl string    `json:"picture_xl"`
	Tracks    resTracks `json:"tracks"`
}

type resPing struct {
	Error   []string `json:"error"`
	Results struct {
		Session         string `json:"SESSION"`
		UserId          int    `json:"USER_ID"`
		Checkform       string `json:"CHECKFORM"`
		ServerTimestamp int    `json:"CHECKFORM"`
	} `json:"results"`
}

var lastReqTime int64

var REQ_MIN_INTERVAL int64 = 500000000

func getConfig() (configuration, error) {
	var err error
	var config configuration

	configDir := os.Getenv("XDG_CONFIG_HOME")
	if len(configDir) == 0 {
		homedir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		configDir = homedir + "/.config/"
	}
	configPath := configDir + "/deezer-flac-download/config.toml"

	_, err = toml.DecodeFile(configPath, &config)
	if err != nil {
		return configuration{}, err
	}
	if len(config.Arl) == 0 {
		return configuration{}, errors.New("please provide a value for the 'arl' field in the config file")
	}
	if len(config.LicenseToken) == 0 {
		return configuration{}, errors.New("please provide a value for the 'license_token' field in the config file")
	}
	if len(config.DestDir) == 0 {
		return configuration{}, errors.New("please provide a value for the 'dest_dir' field in the config file")
	}
	if len(config.PreKey) == 0 {
		return configuration{}, errors.New("please provide a value for the 'pre_key' field in the config file")
	}
	if len(config.Iv) == 0 {
		return configuration{}, errors.New("please provide a value for the 'iv' field in the config file")
	}
	return config, nil
}

func makeReq(method, url string, body io.Reader, config configuration) (*http.Response, error) {
	var err error

	tDiff := time.Now().UnixNano() - lastReqTime
	if tDiff < REQ_MIN_INTERVAL {
		time.Sleep(time.Duration(REQ_MIN_INTERVAL-tDiff) * time.Nanosecond)
	}
	lastReqTime = time.Now().UnixNano()

	// reduce verbosity: do not log every request URL
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Origin", "https://www.deezer.com")
	req.Header.Add("Accept-Language", "en-US,en;q=0.9")
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Referer", "https://www.deezer.com/")
	req.Header.Add("DNT", "1")
	cookie := &http.Cookie{
		Name:  "arl",
		Value: config.Arl,
	}
	req.AddCookie(cookie)

	var res *http.Response
	res, err = http.DefaultClient.Do(req)
	for err != nil {
		log.Print("(network hiccup)")
		res, err = http.DefaultClient.Do(req)
	}
	return res, err
}

func getFavorites(userId string, config configuration) (resTracks, error) {
	url := fmt.Sprintf("https://api.deezer.com/user/%s/tracks?limit=10000000000", userId)
	res, err := makeReq("GET", url, nil, config)
	if err != nil {
		return resTracks{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		bytes, _ := io.ReadAll(res.Body)
		bstr := string(bytes)
		if len(bstr) > 200 {
			bstr = bstr[:200] + "..."
		}
		log.Printf("non-200 response body (truncated): %s", bstr)
		return resTracks{}, fmt.Errorf("got status code %d", res.StatusCode)
	}

	var tracks resTracks
	err = json.NewDecoder(res.Body).Decode(&tracks)
	return tracks, err
}

func getSongInfo(id int64, config configuration) (resSongInfo, error) {
	url := fmt.Sprintf("https://www.deezer.com/de/track/%d", id)

	res, err := makeReq("GET", url, nil, config)
	if err != nil {
		return resSongInfo{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		bytes, _ := io.ReadAll(res.Body)
		bstr := string(bytes)
		if len(bstr) > 200 {
			bstr = bstr[:200] + "..."
		}
		log.Printf("non-200 response body (truncated): %s", bstr)
		return resSongInfo{}, fmt.Errorf("got status code %d", res.StatusCode)
	}

	bytes, _ := io.ReadAll(res.Body)
	s := string(bytes)

	startMarker := `window.__DZR_APP_STATE__ = `
	endMarker := `</script>`
	startIdx := strings.Index(s, startMarker)
	endIdx := strings.Index(s[startIdx:], endMarker)
	sData := s[startIdx+len(startMarker) : startIdx+endIdx]

	var songInfo resSongInfo
	err = json.NewDecoder(strings.NewReader(sData)).Decode(&songInfo)
	return songInfo, err
}

func getAlbum(albumId string, config configuration) (resAlbum, error) {
	url := fmt.Sprintf("https://api.deezer.com/album/%s", albumId)
	res, err := makeReq("GET", url, nil, config)
	if err != nil {
		return resAlbum{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		bytes, _ := io.ReadAll(res.Body)
		bstr := string(bytes)
		if len(bstr) > 200 {
			bstr = bstr[:200] + "..."
		}
		log.Printf("non-200 response body (truncated): %s", bstr)
		return resAlbum{}, fmt.Errorf("got status code %d", res.StatusCode)
	}

	var album resAlbum
	err = json.NewDecoder(res.Body).Decode(&album)
	return album, err
}

func getAlbumSongs(albumId string, config configuration) (resAlbumInfo, error) {
	url := fmt.Sprintf("https://www.deezer.com/de/album/%s", albumId)

	res, err := makeReq("GET", url, nil, config)
	if err != nil {
		return resAlbumInfo{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		bytes, _ := io.ReadAll(res.Body)
		bstr := string(bytes)
		if len(bstr) > 200 {
			bstr = bstr[:200] + "..."
		}
		log.Printf("non-200 response body (truncated): %s", bstr)
		return resAlbumInfo{}, fmt.Errorf("got status code %d", res.StatusCode)
	}

	bytes, _ := io.ReadAll(res.Body)
	s := string(bytes)

	startMarker := `window.__DZR_APP_STATE__ = `
	endMarker := `</script>`
	startIdx := strings.Index(s, startMarker)
	endIdx := strings.Index(s[startIdx:], endMarker)
	sData := s[startIdx+len(startMarker) : startIdx+endIdx]

	var albumInfo resAlbumInfo
	err = json.NewDecoder(strings.NewReader(sData)).Decode(&albumInfo)
	// Ignore error, because we're only unmarshaling SONGS
	return albumInfo, nil
}

// getPlaylist fetches playlist metadata and its full track list.
func getPlaylist(playlistId string, config configuration) (resPlaylist, error) {
	// Fetch playlist page like albums to reuse same ARL cookie and parsing
	var playlist resPlaylist
	apiUrl := fmt.Sprintf("https://api.deezer.com/playlist/%s?access_token=%s", playlistId, config.LicenseToken)
	res, err := makeReq("GET", apiUrl, nil, config)
	if err != nil {
		return resPlaylist{}, err
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		bodyBytes, _ := io.ReadAll(res.Body)
		if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&playlist); err == nil {
			// if API returned tracks, return it
			if playlist.Tracks.Total > 0 || len(playlist.Tracks.Data) > 0 {
				return playlist, nil
			}
		}
		// fall through to webpage parsing below
	}
	// Fallback: fetch playlist page like albums to reuse ARL cookie and parsing
	pageUrl := fmt.Sprintf("https://www.deezer.com/de/playlist/%s", playlistId)
	resPage, err := makeReq("GET", pageUrl, nil, config)
	if err != nil {
		return resPlaylist{}, err
	}
	defer resPage.Body.Close()

	if resPage.StatusCode == 200 {
		bodyBytes, _ := io.ReadAll(resPage.Body)
		s := string(bodyBytes)

		startMarker := `window.__DZR_APP_STATE__ = `
		endMarker := `</script>`
		startIdx := strings.Index(s, startMarker)
		if startIdx >= 0 {
			endIdx := strings.Index(s[startIdx:], endMarker)
			if endIdx >= 0 {
				sData := s[startIdx+len(startMarker) : startIdx+endIdx]
				var generic interface{}
				if err := json.NewDecoder(strings.NewReader(sData)).Decode(&generic); err == nil {
					// try to find playlist title in parsed state
					var foundTitle string
					var walkTitle func(interface{}) bool
					walkTitle = func(n interface{}) bool {
						switch v := n.(type) {
						case map[string]interface{}:
							// common keys: "PLAYLIST" or objects with "TITLE"
							if t, ok := v["PLAYLIST"]; ok {
								if mp, ok2 := t.(map[string]interface{}); ok2 {
									if title, ok3 := mp["TITLE"].(string); ok3 {
										foundTitle = title
										return true
									}
									if title, ok3 := mp["title"].(string); ok3 {
										foundTitle = title
										return true
									}
								}
							}
							if title, ok := v["TITLE"].(string); ok && foundTitle == "" {
								foundTitle = title
								return true
							}
							if title, ok := v["title"].(string); ok && foundTitle == "" {
								foundTitle = title
								return true
							}
							for _, val := range v {
								if walkTitle(val) {
									return true
								}
							}
						case []interface{}:
							for _, el := range v {
								if walkTitle(el) {
									return true
								}
							}
						}
						return false
					}
					walkTitle(generic)
					playlist.Title = foundTitle
				}
			}
		}
		// get tracks using the same page parsing approach
		tracks, err2 := getPlaylistSongs(playlistId, config)
		if err2 == nil {
			playlist.Tracks = tracks
		}
		return playlist, nil
	}

	// get tracks using the same page parsing approach
	tracks, err := getPlaylistSongs(playlistId, config)
	if err == nil {
		playlist.Tracks = tracks
	}

	return playlist, nil
}

// getPlaylistSongs parses the public playlist page and extracts track list
func getPlaylistSongs(playlistId string, config configuration) (resTracks, error) {
	url := fmt.Sprintf("https://www.deezer.com/playlist/%s", playlistId)
	res, err := makeReq("GET", url, nil, config)
	if err != nil {
		return resTracks{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		bytes, _ := io.ReadAll(res.Body)
		bstr := string(bytes)
		if len(bstr) > 200 {
			bstr = bstr[:200] + "..."
		}
		log.Printf("non-200 playlist page body (truncated): %s", bstr)
		return resTracks{}, fmt.Errorf("got status code %d", res.StatusCode)
	}

	bytesBody, _ := io.ReadAll(res.Body)
	s := string(bytesBody)

	startMarker := `window.__DZR_APP_STATE__ = `
	endMarker := `</script>`
	startIdx := strings.Index(s, startMarker)
	if startIdx < 0 {
		return resTracks{}, fmt.Errorf("could not find app state in playlist page")
	}
	endIdx := strings.Index(s[startIdx:], endMarker)
	if endIdx < 0 {
		return resTracks{}, fmt.Errorf("could not find script end in playlist page")
	}
	sData := s[startIdx+len(startMarker) : startIdx+endIdx]

	var generic interface{}
	if err := json.NewDecoder(strings.NewReader(sData)).Decode(&generic); err != nil {
		return resTracks{}, err
	}

	// recursive search for an array of track-like objects
	var found []interface{}
	var walk func(interface{}) bool
	walk = func(n interface{}) bool {
		switch v := n.(type) {
		case map[string]interface{}:
			for _, val := range v {
				if walk(val) {
					return true
				}
			}
		case []interface{}:
			if len(v) > 0 {
				if first, ok := v[0].(map[string]interface{}); ok {
					// heuristic: track object should have an "id" or "SNG_ID"
					if _, hasId := first["id"]; hasId {
						found = v
						return true
					}
					if _, hasSng := first["SNG_ID"]; hasSng {
						// album song info uses SNG_ID, convert later
						found = v
						return true
					}
				}
			}
			// continue search inside array elements
			for _, elem := range v {
				if walk(elem) {
					return true
				}
			}
		}
		return false
	}

	walk(generic)
	if found == nil {
		return resTracks{}, fmt.Errorf("no track array found in playlist page")
	}

	// Convert found array into []resTrack robustly (tolerate type variations)
	tracks := make([]resTrack, 0, len(found))
	for _, el := range found {
		m, ok := el.(map[string]interface{})
		if !ok {
			continue
		}
		var t resTrack

		// id (float64 or string)
		if v, ok := m["id"]; ok {
			switch vv := v.(type) {
			case float64:
				t.Id = int64(vv)
			case string:
				if idInt, err := strconv.ParseInt(vv, 10, 64); err == nil {
					t.Id = idInt
				}
			}
		} else if v, ok := m["SNG_ID"]; ok {
			if s, ok2 := v.(string); ok2 {
				if idInt, err := strconv.ParseInt(s, 10, 64); err == nil {
					t.Id = idInt
				}
			}
		}

		// title
		if v, ok := m["title"]; ok {
			if s, ok2 := v.(string); ok2 {
				t.Title = s
			}
		} else if v, ok := m["SNG_TITLE"]; ok {
			if s, ok2 := v.(string); ok2 {
				t.Title = s
			}
		} else if v, ok := m["SngTitle"]; ok {
			if s, ok2 := v.(string); ok2 {
				t.Title = s
			}
		}

		// md5 image
		if v, ok := m["md5_image"]; ok {
			if s, ok2 := v.(string); ok2 {
				t.Md5Image = s
			}
		} else if v, ok := m["MD5_ORIGIN"]; ok {
			if s, ok2 := v.(string); ok2 {
				t.Md5Image = s
			}
		}

		// album title/md5
		if alb, ok := m["album"]; ok {
			if amap, ok2 := alb.(map[string]interface{}); ok2 {
				if at, ok3 := amap["title"]; ok3 {
					if s, ok4 := at.(string); ok4 {
						t.Album.Title = s
					}
				}
				if md, ok3 := amap["md5_image"]; ok3 {
					if s, ok4 := md.(string); ok4 {
						t.Album.Md5Image = s
					}
				}
			}
		} else {
			if v, ok := m["ALB_TITLE"]; ok {
				if s, ok2 := v.(string); ok2 {
					t.Album.Title = s
				}
			}
		}

		// duration (number or string)
		if v, ok := m["duration"]; ok {
			switch vv := v.(type) {
			case float64:
				t.Duration = int(vv)
			case string:
				if d, err := strconv.Atoi(vv); err == nil {
					t.Duration = d
				}
			}
		} else if v, ok := m["DURATION"]; ok {
			if s, ok2 := v.(string); ok2 {
				if d, err := strconv.Atoi(s); err == nil {
					t.Duration = d
				}
			}
		}

		// best-effort artist
		if art, ok := m["artist"]; ok {
			if amap, ok2 := art.(map[string]interface{}); ok2 {
				if idv, ok3 := amap["id"]; ok3 {
					switch iv := idv.(type) {
					case float64:
						t.Artist.Id = int64(iv)
					case string:
						if idInt, err := strconv.ParseInt(iv, 10, 64); err == nil {
							t.Artist.Id = idInt
						}
					}
				}
				if namev, ok3 := amap["name"]; ok3 {
					if s, ok4 := namev.(string); ok4 {
						t.Artist.Name = s
					}
				}
			}
		}

		tracks = append(tracks, t)
	}

	return resTracks{Data: tracks, Total: len(tracks)}, nil
}

func getSongUrlData(trackToken string, format string, config configuration) (resSongUrl, error) {
	url := "https://media.deezer.com/v1/get_url"
	bodyJsonStr := fmt.Sprintf(`{"license_token":"%s","media":[{"type":"FULL","formats":[{"cipher":"BF_CBC_STRIPE","format":"%s"}]}],"track_tokens":["%s"]}`, config.LicenseToken, format, trackToken)
	res, err := makeReq("POST", url, bytes.NewBuffer([]byte(bodyJsonStr)), config)
	if err != nil {
		return resSongUrl{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		bytes, _ := io.ReadAll(res.Body)
		bstr := string(bytes)
		if len(bstr) > 200 {
			bstr = bstr[:200] + "..."
		}
		log.Printf("non-200 get_url response (truncated): %s", bstr)
		return resSongUrl{}, fmt.Errorf("got status code %d", res.StatusCode)
	}

	var songUrlData resSongUrl
	err = json.NewDecoder(res.Body).Decode(&songUrlData)

	if len(songUrlData.Data) == 0 {
		return resSongUrl{}, fmt.Errorf("got empty Data array when trying to get song URL")
	}

	if len(songUrlData.Data[0].Errors) > 0 {
		return resSongUrl{}, fmt.Errorf("got error when trying to get song URL: %s", songUrlData.Data[0].Errors[0].Message)
	}

	// If Data exists but Media is empty, treat it as "format not available"
	if len(songUrlData.Data[0].Media) == 0 {
		return resSongUrl{}, fmt.Errorf("no media available for requested format %s", format)
	}
	return songUrlData, err
}

func getPing(config configuration) (resPing, error) {
	url := "https://www.deezer.com/ajax/gw-light.php?method=deezer.ping&input=3&api_version=1.0&api_token"
	res, err := makeReq("GET", url, nil, config)
	if err != nil {
		return resPing{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		bytes, _ := io.ReadAll(res.Body)
		bstr := string(bytes)
		if len(bstr) > 200 {
			bstr = bstr[:200] + "..."
		}
		log.Printf("non-200 ping response (truncated): %s", bstr)
		return resPing{}, fmt.Errorf("got status code %d", res.StatusCode)
	}

	var ping resPing
	err = json.NewDecoder(res.Body).Decode(&ping)
	return ping, err
}

func getSongUrl(songUrlData resSongUrl) (string, error) {
	if len(songUrlData.Data) == 0 || len(songUrlData.Data[0].Media) == 0 {
		return "", errors.New("no media available in songUrlData")
	}
	sources := songUrlData.Data[0].Media[0].Sources
	for _, source := range sources {
		if source.Provider == "ak" {
			return source.Url, nil
		}
	}
	return sources[0].Url, nil
}

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
		// basic sanity check: first 4 chars are digits
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
	// Decide extension and folder label based on format
	ext := "flac"
	label := "[WEB FLAC]"
	if strings.HasPrefix(strings.ToUpper(format), "MP3") {
		ext = "mp3"
		label = "[WEB MP3]"
	}
	return fmt.Sprintf("%s/%s/%s - %s %s/%02d - %s.%s", config.DestDir,
		cleanArtist, cleanArtist, cleanAlbumTitle, label, trackNum, cleanSongTitle, ext)
}

func calcBfKey(songId []byte, config configuration) []byte {
	preKey := []byte(config.PreKey)
	songIdHash := md5.Sum(songId)
	songIdMd5 := hex.EncodeToString(songIdHash[:])
	key := make([]byte, 16)
	for i := 0; i < 16; i++ {
		key[i] = songIdMd5[i] ^ songIdMd5[i+16] ^ preKey[i]
	}
	return key
}

func blowfishDecrypt(data []byte, key []byte, config configuration) ([]byte, error) {
	iv, err := hex.DecodeString(config.Iv)
	if err != nil {
		return nil, err
	}
	c, err := blowfish.NewCipher(key)
	if err != nil {
		return nil, err
	}
	cbc := cipher.NewCBCDecrypter(c, iv)
	res := make([]byte, len(data))
	cbc.CryptBlocks(res, data)
	return res, nil
}

func ensureSongDirectoryExists(songPath string, coverUrl string) error {
	var err error
	songDir := path.Dir(songPath)
	if _, err = os.Stat(songDir); errors.Is(err, os.ErrNotExist) {
		os.MkdirAll(songDir, os.ModePerm)

		textFilePath := songDir + "/info.txt"
		textFileData := []byte("Downloaded from Deezer.\n")
		err = os.WriteFile(textFilePath, textFileData, 0644)
		if err != nil {
			return err
		}

		if len(coverUrl) == 0 {
			log.Println("Skipping cover")
		} else {
			coverFilePath := songDir + "/cover.jpg"
			f, err := os.Create(coverFilePath)
			if err != nil {
				return err
			}
			defer f.Close()
			res, err := http.Get(coverUrl)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			if res.StatusCode != 200 {
				return fmt.Errorf("error downloading cover: status %d", res.StatusCode)
			}
			_, err = io.Copy(f, res.Body)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func downloadSong(url string, songPath string, songId string, attempt int, config configuration) error {
	var err error

	if attempt >= 10 {
		return fmt.Errorf("giving up downloading song after %d attempts\n", attempt)
	}

	f, err := os.Create(songPath)
	if err != nil {
		return err
	}
	defer f.Close()

	res, err := makeReq("GET", url, nil, config)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		bytes, _ := io.ReadAll(res.Body)
		bstr := string(bytes)
		if len(bstr) > 200 {
			bstr = bstr[:200] + "..."
		}
		log.Printf("non-200 download response (truncated): %s", bstr)
		return fmt.Errorf("got status code %d", res.StatusCode)
	}

	bfKey := calcBfKey([]byte(songId), config)

	// One in every third 2048 byte block is encrypted
	blockSize := 2048
	buf := make([]byte, blockSize)
	i := 0
	nRead := 0
	totalBytes := 0
	breakNextTime := false

outer_loop:
	for {
		nRead = 0
		for nRead < blockSize {
			nNewRead, err := res.Body.Read(buf[nRead:])
			nRead += nNewRead
			totalBytes += nNewRead
			if breakNextTime {
				break outer_loop
			}
			if err == io.EOF {
				breakNextTime = true
				break
			}
			if err != nil && err != io.EOF {
				log.Printf("Error reading body on i=%d: %s\n", i, err)
				log.Println("Retrying")
				return downloadSong(url, songPath, songId, attempt+1, config)
			}
		}

		isEncrypted := ((i % 3) == 0)
		isWholeBlock := (nRead == blockSize)

		if isEncrypted && isWholeBlock {
			decBuf, err := blowfishDecrypt(buf, bfKey, config)
			if err != nil {
				return fmt.Errorf("error decrypting: %s\n", err)
			}
			f.Write(decBuf)
		} else {
			f.Write(buf[:nRead])
		}

		i += 1
	}

	log.Printf("Wrote %d bytes: %s", totalBytes, songPath)

	return nil
}

func extractFlacComment(f *flac.File) (*flacvorbis.MetaDataBlockVorbisComment, int, error) {
	var err error
	var cmt *flacvorbis.MetaDataBlockVorbisComment
	var cmtIdx int
	for idx, meta := range f.Meta {
		if meta.Type == flac.VorbisComment {
			cmt, err = flacvorbis.ParseFromMetaDataBlock(*meta)
			cmtIdx = idx
			if err != nil {
				return nil, 0, err
			}
		}
	}
	return cmt, cmtIdx, nil
}

func addCover(songPath string, coverPath string) error {
	coverData, err := os.ReadFile(coverPath)
	if err != nil {
		return err
	}

	f, err := flac.ParseFile(songPath)
	if err != nil {
		return err
	}

	picture, err := flacpicture.NewFromImageData(flacpicture.PictureTypeFrontCover,
		"Front cover", coverData, "image/jpeg")
	if err != nil {
		return err
	}

	picturemeta := picture.Marshal()
	f.Meta = append(f.Meta, &picturemeta)
	f.Save(songPath)
	return nil
}

// addID3Tags writes ID3v2 tags and embedded cover to an MP3 file.
func addID3Tags(song resSongInfoData, mp3Path string, coverPath string, album resAlbum) error {
	var tag *id3v2.Tag
	var err error

	tag, err = id3v2.Open(mp3Path, id3v2.Options{Parse: true})
	if err != nil {
		tag = id3v2.NewEmptyTag()
	}
	defer tag.Close()

	title := getTitle(song)
	artist := getArtist(song)
	composer := getComposer(song)

	genre := getAlbumGenres(album)

	tag.SetTitle(title)
	tag.SetAlbum(song.AlbTitle)
	tag.SetArtist(artist)
	tag.AddTextFrame(tag.CommonID("Album artist"), tag.DefaultEncoding(), album.Artist.Name)
	if composer != "" {
		tag.AddTextFrame(tag.CommonID("Composer"), tag.DefaultEncoding(), composer)
	}
	if genre != "" {
		tag.AddTextFrame(tag.CommonID("Content type"), tag.DefaultEncoding(), genre)
	}
	if song.TrackNumber != "" {
		tag.AddTextFrame(tag.CommonID("Track number/Position in set"), tag.DefaultEncoding(), song.TrackNumber)
	}
	if song.DiskNumber != "" {
		tag.AddTextFrame(tag.CommonID("Part of a set"), tag.DefaultEncoding(), song.DiskNumber)
	}
	if song.Copyright != "" {
		tag.AddTextFrame(tag.CommonID("Copyright message"), tag.DefaultEncoding(), song.Copyright)
	}
	// Prefer album release year, fallback to song physical release date
	year := extractYear(album.ReleaseDate)
	if year == "" {
		year = extractYear(song.PhysicalReleaseDate)
	}
	if year != "" {
		tag.SetYear(year)
	}
	if song.Isrc != "" {
		tag.AddTextFrame("TSRC", tag.DefaultEncoding(), song.Isrc)
	}

	if _, err := os.Stat(coverPath); err == nil {
		picBytes, err := os.ReadFile(coverPath)
		if err == nil {
			pf := id3v2.PictureFrame{
				Encoding:    tag.DefaultEncoding(),
				MimeType:    "image/jpeg",
				PictureType: id3v2.PTFrontCover,
				Description: "Cover",
				Picture:     picBytes,
			}
			tag.AddAttachedPicture(pf)
		}
	}

	if err := tag.Save(); err != nil {
		return err
	}
	return nil
}

func addTags(song resSongInfoData, path string, album resAlbum) error {
	var err error

	f, err := flac.ParseFile(path)
	if err != nil {
		return err
	}

	cmts, idx, err := extractFlacComment(f)
	if err != nil {
		return err
	}
	if cmts == nil && idx > 0 {
		cmts = flacvorbis.New()
	}

	title := getTitle(song)
	artist := getArtist(song)
	composer := getComposer(song)

	cmts.Add("TITLE", title)
	cmts.Add("ALBUM", song.AlbTitle)
	cmts.Add("ARTIST", artist)
	cmts.Add("ALBUMARTIST", album.Artist.Name)
	cmts.Add("COMPOSER", composer)
	cmts.Add("TRACKNUMBER", song.TrackNumber)
	cmts.Add("DISCNUMBER", song.DiskNumber)
	cmts.Add("COPYRIGHT", song.Copyright)
	// Add genre (from album) to Vorbis comments
	genre := getAlbumGenres(album)
	if genre != "" {
		cmts.Add("GENRE", genre)
	}
	// Prefer album release year, fallback to song physical release date
	year := extractYear(album.ReleaseDate)
	if year == "" {
		year = extractYear(song.PhysicalReleaseDate)
	}
	if year != "" {
		cmts.Add("DATE", year)
	} else {
		// keep original value if no year could be extracted
		cmts.Add("DATE", song.PhysicalReleaseDate)
	}
	cmts.Add("ISRC", song.Isrc)
	cmtsmeta := cmts.Marshal()
	if idx > 0 {
		f.Meta[idx] = &cmtsmeta
	} else {
		f.Meta = append(f.Meta, &cmtsmeta)
	}

	f.Save(path)

	return nil
}

func printUsage() {
	log.Println("deezer-flac-download is a program to freely download Deezer FLAC files.")
	log.Println("")
	log.Println("To download one or more albums:")
	log.Println("\tdeezer-flac-download album <album_id> [<album_id>...]")
	log.Println("")
	log.Println("To download one or more playlists:")
	log.Println("\tdeezer-flac-download playlist <playlist_id> [<playlist_id>...]")
	log.Println("")
	log.Println("See README for full details.")
}

func main() {
	var err error
	log.SetFlags(0)

	if len(os.Args) < 3 {
		printUsage()
		return
	}

	command := os.Args[1]
	args := os.Args[2:]

	logFilePath := os.TempDir() + "/deezer-flac-download.log"
	logFile, err := os.Create(logFilePath)
	if err != nil {
		log.Fatalf("error creating log file %s: %s\n", logFilePath, err)
	}
	defer logFile.Close()

	config, err := getConfig()
	if err != nil {
		log.Fatalf("error reading config file: %s\n", err)
	}

	if command == "album" {
	album_loop:
		for idx, albumId := range args {
			log.Printf("[%03d/%03d] Downloading album %s\n", idx+1, len(args), albumId)
			albumInfo, err := getAlbumSongs(albumId, config)
			if err != nil {
				log.Fatalf("error getting album songs: %s\n", err)
			}

			album, err := getAlbum(albumId, config)
			if err != nil {
				log.Fatalf("error getting album: %s\n", err)
			}

			for _, song := range albumInfo.Songs.Data {
				// Try FLAC first, then fall back to MP3 variants
				formats := []string{"FLAC", "MP3_320", "MP3_256", "MP3_128"}
				var selectedFormat string
				var songUrl string
				for _, f := range formats {
					songUrlDataTry, errTry := getSongUrlData(song.TrackToken, f, config)
					if errTry != nil {
						continue
					}
					songUrlTry, errTry2 := getSongUrl(songUrlDataTry)
					if errTry2 != nil {
						continue
					}
					selectedFormat = f
					songUrl = songUrlTry
					break
				}

				if selectedFormat == "" {
					msg := fmt.Sprintf("error getting URL for song \"%s\" by %s from \"%s\": no available formats\n",
						song.SngTitle, song.ArtName, song.AlbTitle)
					log.Print(msg)
					logFile.Write([]byte(msg))
					log.Print("Album download failed: " + albumId + "\n\n")
					logFile.Write([]byte("Album download failed: " + albumId + "\n"))
					continue album_loop
				}

				songPath := getSongPath(song, album, config, selectedFormat)
				songDir := path.Dir(songPath)
				coverFilePath := songDir + "/cover.jpg"

				err = ensureSongDirectoryExists(songPath, album.CoverXl)
				if err != nil {
					log.Fatalf("error preparing directory for song: %s\n", err)
				}
				err = downloadSong(songUrl, songPath, song.SngId, 0, config)
				if err != nil {
					log.Fatalf("error downloading song: %s\n", err)
				}

				// Only add FLAC-specific tags/covers when we actually downloaded FLAC
				if strings.ToUpper(selectedFormat) == "FLAC" {
					err = addTags(song, songPath, album)
					if err != nil {
						log.Fatalf("error adding tags to song: %s\n", err)
					}
					err = addCover(songPath, coverFilePath)
					if err != nil {
						log.Fatalf("error adding cover image to song: %s\n", err)
					}
				} else {
					// For MP3 formats, write ID3 tags and attach cover if available
					err = addID3Tags(song, songPath, coverFilePath, album)
					if err != nil {
						log.Fatalf("error adding ID3 tags to MP3: %s\n", err)
					}
					log.Printf("Downloaded %s as %s and added ID3 tags", song.SngTitle, selectedFormat)
				}
			}
			log.Print("Album download succeeded: " + albumId + "\n\n")
			logFile.Write([]byte("Album download succeeded: " + albumId + "\n"))
			log.Print("Album download succeeded: " + albumId + "\n\n")
			logFile.Write([]byte("Album download succeeded: " + albumId + "\n"))
		}
	} else if command == "playlist" {
	playlist_loop:
		for idx, playlistId := range args {
			log.Printf("[%03d/%03d] Downloading playlist %s\n", idx+1, len(args), playlistId)
			playlist, err := getPlaylist(playlistId, config)
			if err != nil {
				log.Fatalf("error getting playlist: %s\n", err)
			}

			// Try to obtain tracks via API; if empty, parse playlist page
			tracks := playlist.Tracks
			if tracks.Total == 0 || len(tracks.Data) == 0 {
				tracksParsed, err2 := getPlaylistSongs(playlistId, config)
				if err2 == nil {
					tracks = tracksParsed
				} else {
					log.Printf("could not extract playlist tracks from page: %v\n", err2)
				}
			}

			for _, track := range tracks.Data {
				songInfo, err := getSongInfo(track.Id, config)
				if err != nil {
					log.Fatalf("error getting song info: %s\n", err)
				}
				song := songInfo.Data

				album, err := getAlbum(song.AlbId, config)
				if err != nil {
					log.Fatalf("error getting album: %s\n", err)
				}

				// Try FLAC first, then fall back to MP3 variants
				formats := []string{"FLAC", "MP3_320", "MP3_256", "MP3_128"}
				var selectedFormat string
				var songUrl string
				for _, f := range formats {
					songUrlDataTry, errTry := getSongUrlData(song.TrackToken, f, config)
					if errTry != nil {
						continue
					}
					songUrlTry, errTry2 := getSongUrl(songUrlDataTry)
					if errTry2 != nil {
						continue
					}
					selectedFormat = f
					songUrl = songUrlTry
					break
				}

				if selectedFormat == "" {
					msg := fmt.Sprintf("error getting URL for song \"%s\" by %s from \"%s\": no available formats\n",
						song.SngTitle, song.ArtName, song.AlbTitle)
					log.Print(msg)
					logFile.Write([]byte(msg))
					log.Print("Playlist download failed: " + playlistId + "\n\n")
					logFile.Write([]byte("Playlist download failed: " + playlistId + "\n"))
					continue playlist_loop
				}

				songPath := getSongPath(song, album, config, selectedFormat)
				songDir := path.Dir(songPath)
				coverFilePath := songDir + "/cover.jpg"

				err = ensureSongDirectoryExists(songPath, album.CoverXl)
				if err != nil {
					log.Fatalf("error preparing directory for song: %s\n", err)
				}
				err = downloadSong(songUrl, songPath, song.SngId, 0, config)
				if err != nil {
					log.Fatalf("error downloading song: %s\n", err)
				}

				if strings.ToUpper(selectedFormat) == "FLAC" {
					err = addTags(song, songPath, album)
					if err != nil {
						log.Fatalf("error adding tags to song: %s\n", err)
					}
					err = addCover(songPath, coverFilePath)
					if err != nil {
						log.Fatalf("error adding cover image to song: %s\n", err)
					}
				} else {
					err = addID3Tags(song, songPath, coverFilePath, album)
					if err != nil {
						log.Fatalf("error adding ID3 tags to MP3: %s\n", err)
					}
					log.Printf("Downloaded %s as %s and added ID3 tags", song.SngTitle, selectedFormat)
				}
			}
			log.Print("Playlist download succeeded: " + playlistId + "\n\n")
			logFile.Write([]byte("Playlist download succeeded: " + playlistId + "\n"))
		}
	} else {
		printUsage()
		return
	}
}
