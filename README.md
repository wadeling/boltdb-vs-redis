# boltdb-vs-redis

this is a little performance between boltdb and redis

## test method

- total 2000 key-value data
- first store data, then retrive them

## result

|  software | read | write | rw|
|  ----  | ----  | ---- | -----|
| bolt  | 6ms | 45s | 45s |
| redis  | 3s | 3s | 7s |

boltdb cost most time in writing, reading is fast

