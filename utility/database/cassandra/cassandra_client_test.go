// Copyright 2015 CloudAwan LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cassandra

/*
import (
	"fmt"
	"testing"
	"time"
)


var tableSchema = `
CREATE TABLE IF NOT EXISTS test_table (
column1 text,
column2 text,
column3 int,
PRIMARY KEY (column1, column2, column3));
`

func TestCreateTable(t *testing.T) {
	session, err := CassandraClient.GetSession()
	fmt.Println(session, err)

	if err := session.Query(tableSchema).Exec(); err != nil {
		t.Errorf("Check if not exist then create table error: %s", err)
	}
}


func TestInsertData(t *testing.T) {
	session, err := CassandraClient.GetSession()
	fmt.Println(session, err)

	for i := 0; i < 1000; i++ {
		err := session.Query("INSERT INTO test_table (column1, column2, column3) VALUES ('Jones', 'Austin', 35)").Exec()
		fmt.Println(err)
		time.Sleep(time.Second)
	}
}


func TestIterationQueryData(t *testing.T) {
	session, err := CassandraClient.GetSession()
	fmt.Println(session, err)

	var column1, column2 string
	var column3 int

	for i := 0; i < 1000; i++ {
		iter := session.Query("SELECT * FROM test_table").Iter()
		for iter.Scan(&column1, &column2, &column3) {
			fmt.Println(column1, column2, column3)
		}
		time.Sleep(time.Second)
	}

}


func TestScanQueryData(t *testing.T) {
	session, err := CassandraClient.GetSession()
	fmt.Println(session, err)

	var column1, column2 string
	var column3 int

	for i := 0; i < 1000; i++ {
		err := session.Query("SELECT * FROM test_table WHERE column1 = 'Jones'").Scan(&column1, &column2, &column3)
		fmt.Println(err, column1, column2, column3)
		column1 = ""
		column2 = ""
		column3 = 0
		time.Sleep(time.Second)
	}

}
*/
