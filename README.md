# Violet [![Build Status](https://travis-ci.org/cosmtrek/violet.svg?branch=master)](https://travis-ci.org/cosmtrek/violet)

```
        __        ___ ___
\  / | /  \ |    |__   |
 \/  | \__/ |___ |___  |

```
A simple search engine in go.

## Install

```
go get -u github.com/cosmtrek/violet

# dont forget add `violet` environment variable
# vi ~/.bash_profile
export violet='$GOPATH/src/github.com/cosmtrek/violet'
```

There are two modes for violet, one is for guys who loves terminal like me, the another is running as http server.

### Terminal Mode

```
violet -path=INDEX_PATH -index=INDEX_NAME -fields=INDEX_FIELDS -data=DATA_FILE -query=true -server=false
```

Then you can search anything you feed in.

### Server Mode

```
# start server
violet
```

After the server is started, open another terminal to make a post request to create index.

```
# first create a json file post.json
{
    "index": "INDEX_NAME",
    "index_path": "INDEX_PATH",
    "fields": "INDEX_FIELDS",
    "datafile": "DATA_FILE"
}
# then create index
curl -XPOST -d @./data/tweets.json "http://localhost:6060/index"
# try to query
curl "http://localhost:6000/INDEX_NAME/search?query=TERM"
```

## Query

In order to search with efficiency, it's necessary to query something with conditions. Currently only support following
forms:

1. `word` search single word
2. `word1 word2` search multiple words
3. `word1 -word2` search word1 and excludes word2
4. `field1:word1 field2:word2` search word1 in field1 and word2 in field2
5. `word field>10` search word and field(integer) is greater than 10
