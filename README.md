# boltdb-vs-redis

this is a little performance between boltdb and redis

## test method

- total 2000 key-value data
- first store data, then retrive them

## result
- boltdb: 46s
- redis: 5s

boltdb cost most time in writing, reading is fast

