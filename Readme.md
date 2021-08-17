# htab

htab - an html table to csv / tsv / json / yaml converter
---

Install:
---
install the normal "go" way:
```
go get github.com/cbluth/htab
```

Usage:
---

you can pass the URL of the page as an argument or pipe the html data to stdin.

```
# standard usage
htab $URL

# or pass html to stdin
curl -s https://example.com/some-table \
    | htab
```
```
# use comma delimiters (default)
htab $URL

# or explicitly specify using commas delimiters
htab -d',' $URL 
```
```
# use tab delimiters
htab -d'\t' $URL
```
```
# use pipe delimiters
htab -d'|' $URL
```
```
# use semicolon delimiters
htab -d';' $URL
```
```
# use first table on page
htab -n1 $URL
```
```
# use second table on page (or third, forth, etc)
htab -n2 $URL
htab -n3 $URL
```
```
# convert to json
htab -j $URL
```
```
# convert to yaml
htab -y $URL
```
```
# advanced, use third table on page with pipe delimiters
htab -n3 -d'|' $URL
```

---
If you have any issues, please submit them here:
https://github.com/cbluth/htab/issues
