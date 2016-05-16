# IMDB
IMDB scans movie files on disc and tries to find it on www.imdb.com to extract data

## Usage
Create dirs.yml with absolute paths to directories of movie files. IMDB ignores files with underscore at the beginning.

Example:
```
- /movies/
- /home/user/movies/
```

or Windows

```
- c:/movies/
- d:/movies/
```

## Output
movies.txt and movies.html with

```Name Duration Rating #Rating Year FSK IMDB-Link Size File```
