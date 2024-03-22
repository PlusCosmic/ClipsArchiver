# ClipsArchiver
[](https://github.com/PlusCosmic/ClipsArchiver.Server/actions/workflows/go.yml/badge.svg)

[Client application](https://github.com/PlusCosmic/ClipsArchiver.Client.Windows)

Server component for the Clips Archiver project comprised of three services:

### ClipsArchiver:
  - Allows external interaction with the system through a REST API and static filesystem
  - supports uploading gameplay clips
  - hosts clips and thumbnails on a static file system for the client to retrieve
  - hosts image resources for client to retrieve
  - Retrieve transcoding queue
  - Retrieve list of clip objects for a given date
  - Retrieve other information useful to the client including all users, apex map information, apex legend information, all known tags

### ClipsTranscoder:
  - Frequently polls the queue table in the database and transcodes all clips to 1080p
  - Gets information from the file such as video duration
  - Generates video thumbnails
  - Updates database queue entries to keep the client app up to date with the transcode progress

### MatchHistoryProcessor:
  - Polls the Apex Legends Status API to retrieve historic match data
  - Tries to match clips to the match data and fill in extra information on the clip object including the legend played, and the map

## Setup
1. Clone and build the three applications in /cmd/
2. Setup database using script in /DB Scripts/
3. Run any of the applications once to generate config files
4. Populate config files with storage paths, API key for ALS, and database information
5. Run all three applications
