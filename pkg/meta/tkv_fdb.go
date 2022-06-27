/*
 * JuiceFS, Copyright 2021 Juicedata, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package meta

import (
	"fmt"
	"strings"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
)

func init() {
	Register("fdb", newKVMeta)
	drivers["fdb"] = newFdbClient
}

const DEBUG = true

type fdbTxn struct {
	t fdb.Transaction
}

type fdbClient struct {
	client fdb.Database
}

func newFdbClient(address string) (tkvClient, error) {
	args := strings.Split(address, ":")
	err := fdb.APIVersion(610)
	if err != nil {
		fmt.Printf("Unable to set API version: %v\n", err)
		return &tikvClient{}, err
	}
	db, err := fdb.OpenDatabase(args[0])
	prefix := ""
	if len(args) == 2 {
		prefix = args[1]
	}
	return withPrefix(&fdbClient{db}, append([]byte(prefix), 0xFD)), err
}

func (c *fdbClient) name() string {
	return "fdb"
}

func (c *fdbClient) txn(f func(kvTxn) error) error {
	tx, err := c.client.CreateTransaction()
	if err != nil {
		return err
	}
	defer func(e *error) {
		if r := recover(); r != nil {
			fe, ok := r.(error)
			if ok {
				*e = fe
			} else {
				panic(r)
			}
		}
	}(&err)
	if err = f(&fdbTxn{tx}); err != nil {
		tx.Reset()
		return err
	}
	return tx.Commit().Get()
}

func (c *fdbClient) scan(prefix []byte, handler func(key, value []byte)) error {
	tx, err := c.client.CreateTransaction()
	if err != nil {
		return err
	}
	iter := tx.GetRange(
		fdb.KeyRange{Begin: fdb.Key(prefix), End: fdb.Key(nextKey(prefix))},
		fdb.RangeOptions{},
	).Iterator()
	for iter.Advance() {
		r := iter.MustGet()
		handler(r.Key, r.Value)
	}
	//fmt.Println("scan ----------------")
	return nil
}

func (c *fdbClient) reset(prefix []byte) error {
	tx, err := c.client.CreateTransaction()
	if err != nil {
		panic(err)
	}
	tx.ClearRange(fdb.KeyRange{
		Begin: fdb.Key(prefix),
		End:   fdb.Key(nextKey(prefix)),
	})
	//fmt.Println("fdbClient reset !----------------")
	return tx.Commit().Get()
}

func (c *fdbClient) close() error {
	c = &fdbClient{}
	//fmt.Println("fdbClient close !----------------")
	return nil
}

func (c *fdbClient) shouldRetry(err error) bool {
	ret := strings.Contains(err.Error(), "Transaction not committed") || strings.Contains(err.Error(), "Operation aborted")
	/*fmt.Println(err.Error())
	if ret {
		fmt.Println("should retry ")
	}*/
	return ret
}

func (tx *fdbTxn) get(key []byte) []byte {
	//fmt.Println("fdbTxn get !----------------")
	return tx.t.Get(fdb.Key(key)).MustGet()
}

func (tx *fdbTxn) gets(keys ...[]byte) [][]byte {
	ret := make([][]byte, 0)
	for _, key := range keys {
		val := tx.t.Get(fdb.Key(key)).MustGet()
		ret = append(ret, val)
	}
	//fmt.Println("fdbTxn gets !----------------")
	return ret
}
func (tx *fdbTxn) range0(begin, end []byte) fdb.RangeIterator {
	return *(tx.t.GetRange(
		fdb.KeyRange{Begin: fdb.Key(begin), End: fdb.Key(end)},
		fdb.RangeOptions{},
	).Iterator())
}
func (tx *fdbTxn) scanRange(begin, end []byte) map[string][]byte {
	ret := make(map[string][]byte, 0)
	iter := tx.range0(begin, end)
	for iter.Advance() {
		r := iter.MustGet()
		ret[string(r.Key)] = r.Value
	}
	//fmt.Println("fdbTxn scanRange !----------------")
	return ret
}
func (tx *fdbTxn) scanKeys(prefix []byte) [][]byte {
	ret := make([][]byte, 0)
	iter := tx.range0(prefix, nextKey(prefix))
	for iter.Advance() {
		ret = append(ret, iter.MustGet().Key)
	}
	//fmt.Println("fdbTxn scanKey !----------------")
	return ret
}
func (tx *fdbTxn) scanValues(prefix []byte, limit int, filter func(k, v []byte) bool) map[string][]byte {
	ret := make(map[string][]byte, 0)
	iter := tx.range0(prefix, nextKey(prefix))
	for iter.Advance() {
		r := iter.MustGet()
		if filter(r.Key, r.Value) {
			ret[string(r.Key)] = r.Value
		}
	}
	//fmt.Println("fdbTxn scanValue !----------------")
	return ret
}
func (tx *fdbTxn) exist(prefix []byte) bool {
	iter := tx.range0(prefix, nextKey(prefix))
	//fmt.Println("fdbTxn exist !----------------")
	return iter.Advance()
}
func (tx *fdbTxn) set(key, value []byte) {
	//fmt.Println("fdbTxn set !----------------")
	tx.t.Set(fdb.Key(key), fdb.Key(value))
}
func (tx *fdbTxn) append(key []byte, value []byte) []byte {
	tx.t.AppendIfFits(fdb.Key(key), fdb.Key(value))
	//fmt.Println("fdbTxn append !----------------")
	return tx.t.Get(fdb.Key(key)).MustGet()
}
func (tx *fdbTxn) incrBy(key []byte, value int64) int64 {
	new := parseCounter(tx.t.Get(fdb.Key(key)).MustGet())
	if value != 0 {
		new += value
		tx.t.Set(fdb.Key(key), fdb.Key(packCounter(new)))
	}
	//fmt.Println("fdbTxn increaby !----------------")
	return new
}
func (tx *fdbTxn) dels(keys ...[]byte) {
	for _, key := range keys {
		tx.t.Clear(fdb.Key(key))
	}
	//fmt.Println("fdbTxn dels !----------------")
}
