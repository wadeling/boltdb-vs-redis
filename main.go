package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/go-redis/redis/v8"
)

var world = []byte("world")

var (
	keyPrefix       = "hello"
	valuePrefix     = "world"
	num             = 2000
	testData        = make(map[string]string,0)
	db              *bolt.DB
	redisHost       = "redis://localhost:6379"
	redisOption     *redis.Options
	redisClient     *redis.Client
)

func initDb(path string) (err error) {
	db, err = bolt.Open(path, 0644, nil)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func initRedis() (err error) {
	redisOption, err = redis.ParseURL(redisHost)
	if err != nil {
		return err
	}
	redisClient = redis.NewClient(redisOption)
	log.Printf("redis option %+v",redisOption)
	return nil
}

func prepareData() {
	for i := 0;i<num;i++ {
		key := fmt.Sprintf("%s%d",keyPrefix,i)
		value := fmt.Sprintf("%s%d",valuePrefix,i)
		testData[key]= value
	}
	log.Printf("test data count %d",len(testData))
}

func redisBatchWrite() error {
	//set data
	for key,val := range testData {
		//set
		if err := redisClient.Set(context.TODO(), key, val, 0).Err(); err != nil {
			log.Fatalf("redis set err, key %s,val %s,%v",key,val,err)
			return err
		}
	}
	log.Printf("redis save data end")
	return nil
}

func redisBatchRead() error {
	// get data
	for key,val := range testData {
		v, err := redisClient.Get(context.TODO(), key).Bytes()
		if err == redis.Nil {
			return fmt.Errorf("key (%s) is missing ",key)
		} else if err != nil {
			return fmt.Errorf("redis get val err.key %s,err %v",key,err)
		}
		if string(v) != val {
			return fmt.Errorf("val not match,%s ,%s",string(v),val)
		}
		//log.Printf("val %s",string(v))
	}

	return nil
}

func boltdbBatchWrite() error {
	// store some data
	for key, value:= range testData {
		err := db.Update(func(tx *bolt.Tx) error {
			bucket, err := tx.CreateBucketIfNotExists(world)
			if err != nil {
				return err
			}

			err = bucket.Put([]byte(key), []byte(value))
			if err != nil {
				log.Fatalf("bolt put err,%v",err)
				return err
			}
			return nil
		})

		if err != nil {
			log.Fatal(err)
			return err
		}
	}
	log.Printf("store data end")
	return nil
}

func boltdbBatchRead() error {
	// retrieve the data
	for key, value := range testData {
		err := db.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket(world)
			if bucket == nil {
				return fmt.Errorf("bucket %v not found", world)
			}

			val := bucket.Get([]byte(key))
			if string(val) != value {
				fmt.Printf("get wrong val,val %s,value %s",string(val),value)
				return fmt.Errorf("get wrong val")
			}
			//fmt.Println(string(val))

			return nil
		})

		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: %s {bolt|redis} {r|w|rw}",os.Args[0])
		return
	}

	prepareData()

	action := os.Args[1]
	//init
	if action == "bolt" {
		if err := initDb("./bolt.db"); err != nil {
			fmt.Errorf("open db err:%v",err)
			return
		}
		defer db.Close()
	} else if action == "redis" {
		if err := initRedis();err != nil {
			fmt.Errorf("init redis err,%v",err)
			return
		}
	}

	mode := os.Args[2]

	start := time.Now()
	defer func() {
		tc := time.Since(start)
		log.Printf("all scan time cost %v",tc)
	}()

	//start test
	if action == "bolt" {
		log.Println("start bolt perf test")
		if mode == "r" {
			boltdbBatchRead()
		} else if mode == "w" {
			boltdbBatchWrite()
		} else if mode == "rw" {
			boltdbBatchWrite()
			boltdbBatchRead()
		}
	} else if action == "redis" {
		log.Println("start redis perf test")
		if mode == "r" {
			redisBatchRead()
		} else if mode == "w" {
			redisBatchWrite()
		} else if mode == "rw" {
			redisBatchWrite()
			redisBatchRead()
		}
	}

	log.Println("end")
}

