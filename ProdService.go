package MyPractice

import (
	DB "MyPractice/MyDB"
	"context"
	"fmt"
	"github.com/boltdb/bolt"
	"io"
	"log"
	"os"
	"strings"
)

type ProdService struct {}

/*
func (this *ProdService)GetProdStock(context.Context, *ProdRequest) (*ProdResponse, error){
	return &ProdResponse{Response:"yes, I can!"}, nil
}*/

func (this *ProdService)CreateBucket(con context.Context,req *Request) (res *Response, err error){
	res = &Response{RespondResult: false}
	err = nil

	if req.BucketName == "" {
		res.RespondResult = false
		res.ResultStr = "遗漏了 '--name' 请重试."
		return
	}
	var db *bolt.DB
	db, err = DB.DBInit("my.db")
	if err != nil{
		return
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error{
		//TODO: 创建bucket
		fmt.Printf("%s\n",res.ResultStr)
		_, err = tx.CreateBucketIfNotExists([]byte(req.BucketName))
		if err != nil {
			res.RespondResult = false
			res.ResultStr = fmt.Sprintf("bucket创建失败.")
			log.Print(err)
			return err
		}

		DB.MyBuckets[DB.GetBucketNum()] = req.BucketName

		res.RespondResult = true
		res.ResultStr = fmt.Sprintf("bucket创建成功. name : %s", req.BucketName)
		fmt.Printf("%s\n",res.ResultStr)
		return nil
	})

	return
}

func (this *ProdService)DeleteBucket(con context.Context,req *Request) (res *Response, err error){
	res = &Response{RespondResult: false}
	err = nil

	if req.BucketName == "" {
		res.RespondResult = false
		res.ResultStr = "遗漏了 '--name' 请重试."
		return
	}

	var db *bolt.DB
	db, err = DB.DBInit("my.db")
	if err != nil{
		return
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error{
		//TODO: 删除bucket
		for index, name := range DB.MyBuckets {
			if name == req.BucketName{
				err := tx.DeleteBucket([]byte(name))
				if err != nil {
					res.RespondResult = false
					res.ResultStr = "删除失败，请检查."
					return err
				}
				res.RespondResult = true
				res.ResultStr = "删除成功！"
				delete(DB.MyBuckets, index)
				return nil
			}
		}
		res.RespondResult = false
		res.ResultStr = "删除失败，并没有这个bucket."
		return nil
	})
	return
}

func (this *ProdService)SetKey(con context.Context,req *Request) (res *Response, err error){
	res = &Response{RespondResult: false}
	err = nil

	if req.BucketName == "" || req.KeyName == "" || req.ValueName == ""{
		res.RespondResult = false
		res.ResultStr = "请检查是否遗漏了 Flag：--key --value --bucket"
		return
	}

	var db *bolt.DB
	db, err = DB.DBInit("my.db")
	if err != nil{
		return
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error{
		//TODO: 设置key/value的值（指定bucketname后）
		for _, name := range DB.MyBuckets {
			if name == req.BucketName{
				bucket := tx.Bucket([]byte(name))
				err := bucket.Put([]byte(req.KeyName), []byte(req.ValueName))
				if err != nil {
					res.RespondResult = false
					res.ResultStr = fmt.Sprintf("添加键值对 %s 失败", req.KeyName)
					log.Print(err)
					return err
				}
				res.RespondResult = true
				res.ResultStr = fmt.Sprintf("添加键值对 %s 成功", req.KeyName)
			}
		}
		return nil
	})

	return
}

func (this *ProdService)GetKey(con context.Context,req *Request) (res *Response, err error){
	res = &Response{RespondResult: false}
	res.ValuesArray = make([]string, 50)
	var ValueNums int = 0
	err = nil

	if req.BucketName == "" || req.KeyName == "" {
		res.RespondResult = false
		res.ResultStr = "请检查是否遗漏了 Flag：--key --bucket"
		return
	}

	var db *bolt.DB
	db, err = DB.DBInit("my.db")
	if err != nil{
		return
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error{
		//TODO: 根据key值获取value的值，且还需要判断prefix标志位。
		for _, name := range DB.MyBuckets {
			if name == req.BucketName{
				bucket := tx.Bucket([]byte(name))
				if req.Prefix{	//true
					//以map作为缓冲，取出所有bucket所有的key/value， 然后进行prefix判断
					mapBuf := make(map[string]string)
					cursor := bucket.Cursor()
					for k, v := cursor.First(); k != nil; k, v = cursor.Next(){
						mapBuf[string(k)] = string(v)
					}

					//执行prefix判断
					for k, v := range mapBuf{
						if strings.HasPrefix(k, req.KeyName){
							res.ValuesArray[ValueNums] = v
							ValueNums++
						}
					}

					res.RespondResult = true
				}else{	//false
					res.ValueName = string(bucket.Get([]byte(req.KeyName)))
					if res.ValueName != "" {
						res.RespondResult = true
						res.ResultStr = fmt.Sprintf("键 %s 对应的值为：%s", req.KeyName, res.ValueName)
					}else{
						res.RespondResult = false
						res.ResultStr = fmt.Sprintf("数据库中不存在键 %s ，请检查", req.KeyName)
					}
				}
			}
		}
		return nil
	})

	return
}

func (this *ProdService)DelKey(con context.Context,req *Request) (res *Response, err error){
	res = &Response{RespondResult: false}
	err = nil

	if req.BucketName == "" || req.KeyName == "" {
		res.RespondResult = false
		res.ResultStr = "请检查是否遗漏了 Flag：--key --bucket"
		return
	}

	var db *bolt.DB
	db, err = DB.DBInit("my.db")
	if err != nil{
		return
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error{
		//TODO: 根据key删除 键值数据

		bucket := tx.Bucket([]byte(req.BucketName))
		if bucket == nil{
			return  fmt.Errorf("bucket not exist")
		}
		err := bucket.Delete([]byte(req.KeyName))
		if err != nil {
			res.RespondResult = false
			res.ResultStr = fmt.Sprintf("删除键值对 %s 失败", req.KeyName)
			return err
		}
		res.RespondResult = true
		res.ResultStr = fmt.Sprintf("删除键值对 %s 成功", req.KeyName)




		for _, name := range DB.MyBuckets {
			if name == req.BucketName{
				bucket := tx.Bucket([]byte(name))
				err := bucket.Delete([]byte(req.KeyName))
				if err != nil {
					res.RespondResult = false
					res.ResultStr = fmt.Sprintf("删除键值对 %s 失败", req.KeyName)
					return err
				}
				res.RespondResult = true
				res.ResultStr = fmt.Sprintf("删除键值对 %s 成功", req.KeyName)
			}
		}


		bucket := tx.Bucket([]byte(req.BucketName))
		err := bucket.Delete([]byte(req.KeyName))
		if err != nil {
			res.RespondResult = false
			res.ResultStr = fmt.Sprintf("删除键值对 %s 失败", req.KeyName)
			return err
		}
		res.RespondResult = true
		res.ResultStr = fmt.Sprintf("删除键值对 %s 成功", req.KeyName)


		return nil
	})
	return
}

func (this *ProdService)Backup(con context.Context, req *Nothing) (no *Nothing, err error){
	no = new(Nothing)
	_, err = CopyFile("my.db", "my.db.bak")
	return
}

func (this *ProdService)Restore(con context.Context, req *Nothing) (no *Nothing, err error){
	no = new(Nothing)
	_, err = CopyFile("my.db.bak","my.db")
	return
}

func CopyFile(dstName, srcName string) (writeen int64,err error) {
	src, err := os.Open(dstName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer src.Close()

	dst, err := os.OpenFile(srcName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dst.Close()
	return io.Copy(dst, src)
}