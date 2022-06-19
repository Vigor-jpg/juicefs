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
	"github.com/apple/foundationdb/bindings/go/src/fdb"
)

func init() {
	Register("fdb", newKVMeta)
	drivers["fdb"] = newFdbClient
}

type fdbTxn struct {
	t fdb.Transaction
}

type fdbClient struct {
	client fdb.Database
}

func newFdbClient(address string) (tkvClient, error) {
	db, err := fdb.OpenDatabase(address)
	return withPrefix(&fdbClient{db}, []byte("")), err
}

func (c *fdbClient) name() string {
	return "fdb"
}

func (c *fdbClient) txn(f func(kvTxn) error) error {
	tx, err := c.client.CreateTransaction()
	if err != nil {
		return err
	}
	if err = f(&fdbTxn{tx}); err != nil {
		return err
	}
	return tx.Commit().Get()
}

func (c *fdbClient) scan(prefix []byte, handler func(key, value []byte)) error {

}

func (c *fdbClient) reset(prefix []byte) error {

}

func (c *fdbClient) close() error {

}

func (c *fdbClient) shouldRetry(err error) bool {

}

func (tx *fdbTxn) get(key []byte) []byte {

}

func (tx *fdbTxn) gets(keys ...[]byte) [][]byte {

}
func (tx *fdbTxn) scanRange(begin, end []byte) map[string][]byte {

}
func (tx *fdbTxn) scanKeys(prefix []byte) [][]byte {

}
func (tx *fdbTxn) scanValues(prefix []byte, limit int, filter func(k, v []byte) bool) map[string][]byte {

}
func (tx *fdbTxn) exist(prefix []byte) bool {

}
func (tx *fdbTxn) set(key, value []byte) {

}
func (tx *fdbTxn) append(key []byte, value []byte) []byte {

}
func (tx *fdbTxn) incrBy(key []byte, value int64) int64 {

}
func (tx *fdbTxn) dels(keys ...[]byte) {

}
