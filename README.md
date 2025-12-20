# deezer-flac-download

A program to freely download Deezer Audio files. Tested and working in December 2025.
Verified to produce the same audio as other downloaders being used for files present on the internet. A paid Deezer account is required.
 
Note: The tool will now automatically fall back to downloading MP3 files when the FLAC format is not available for a given track. In that case the downloader will try `FLAC` first, then `MP3_320`, `MP3_256`, and finally `MP3_128`. 

The program downloads cover art and metadata tags: for MP3s it writes ID3v2 tags and embeds the cover image into the MP3 file, and for FLACs it embeds the cover art and metadata.

## Setup

Create a file at `~/.config/deezer-flac-download/config.toml` based on
`example_config.toml`. The contents are as follows:

* `arl`: Can be obtained from the `arl` cookie in your browser.
* `license_token`: Navigate to a song page, open the "Network" tab in your
  browser's dev tools, click the play button, click the "get_url" request, find
  the request data in the right sidebar and you'll find the `license_token`
  there.
* `dest_dir`: Choose any folder.
* `pre_key` and `iv`: Fill them in with the values you magically found at https://bin.0xfc.de/?489876949a0c544c#3UYL7DBfD2RjHRjW86BFVFeJJBwrTftop5Lvgrvo3Wsb

## Usage

1. Find the album's ID by navigating to it and looking at the URL. It's the
  string of numbers.
1. `go run . album <album_id>`

You can also download multiple albums: `go run . album 1234 2345 3456`.

## Note

Recent additions to this repository have been made with the help of an AI assistant (Copilot).
