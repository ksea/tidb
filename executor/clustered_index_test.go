// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package executor_test

import (
	. "github.com/pingcap/check"
	"github.com/pingcap/tidb/util/testkit"
)

type testClusteredSuite struct{ *baseTestSuite }

func (s *testClusteredSuite) SetUpTest(c *C) {
}

func (s *testClusteredSuite) newTK(c *C) *testkit.TestKit {
	tk := testkit.NewTestKitWithInit(c, s.store)
	tk.MustExec("set @@tidb_enable_clustered_index = 1")
	return tk
}

func (s *testClusteredSuite) TestClusteredUnionScan(c *C) {
	tk := s.newTK(c)
	tk.MustExec("drop table if exists t")
	tk.MustExec("CREATE TABLE t (a int,b int,c int, PRIMARY KEY (a,b))")
	tk.MustExec("insert t (a, b) values (1, 1)")
	tk.MustExec("begin")
	tk.MustExec("update t set c = 1")
	tk.MustQuery("select * from t").Check(testkit.Rows("1 1 1"))
	tk.MustExec("rollback")

	// cover old row format.
	tk = testkit.NewTestKitWithInit(c, s.store)
	tk.Se.GetSessionVars().RowEncoder.Enable = false
	tk.MustExec("begin")
	tk.MustExec("update t set c = 1")
	tk.MustQuery("select * from t").Check(testkit.Rows("1 1 1"))
	tk.MustExec("rollback")
}

func (s *testClusteredSuite) TestClusteredIndexLookUp(c *C) {
	tk := s.newTK(c)
	tk.MustExec("drop table if exists t")
	tk.MustExec("create table t (a int, b int, c int, d int, primary key (a, b))")
	tk.MustExec("create index idx on t(c)")
	tk.MustExec("insert t values (1, 1, 1, 1)")
	tk.MustQuery("select d from t use index (idx)").Check(testkit.Rows("1"))
}

func (s *testClusteredSuite) TestClusteredIndexLookUp2(c *C) {
	tk := s.newTK(c)
	tk.MustExec("drop table if exists c3")
	createTable := `
CREATE TABLE c3 (
  c_id int(11) NOT NULL,
  c_d_id int(11) NOT NULL,
  c_w_id int(11) NOT NULL,
  c_first varchar(16) DEFAULT NULL,
  c_middle char(2) DEFAULT NULL,
  c_last varchar(16) DEFAULT NULL,
  c_balance decimal(12,2) DEFAULT NULL,
  PRIMARY KEY (c_w_id,c_d_id,c_id),
  KEY idx (c_w_id,c_d_id,c_last,c_first)
);`
	tk.MustExec(createTable)
	tk.MustExec("insert c3 values (772,1,1,'aaa','OE','CALL',0),(1905,1,1,'bbb','OE','CALL',0);")
	query := `
SELECT c_balance, c_first, c_middle, c_id FROM c3 use index (idx) WHERE c_w_id = 1 AND c_d_id = 1 and c_last = 'CALL' ORDER BY c_first
`
	tk.MustQuery(query).Check(testkit.Rows("0.00 aaa OE 772", "0.00 bbb OE 1905"))
}

func (s *testClusteredSuite) TestClusteredTopN(c *C) {
	tk := s.newTK(c)
	tk.MustExec("drop table if exists o3")
	createTables := `
	CREATE TABLE o3 (
	o_id int NOT NULL,
	o_d_id int,
	o_w_id int,
	o_c_id int,
	PRIMARY KEY (o_w_id,o_d_id,o_id),
	KEY idx_order (o_w_id,o_d_id,o_c_id,o_id)
);`
	tk.MustExec(createTables)
	tk.MustExec("insert o3 values (1, 6, 9, 3), (2, 6, 9, 5), (3, 6, 9, 7)")
	tk.MustQuery("SELECT max(o_id) max_order FROM o3 use index (idx_order)").Check(testkit.Rows("3"))
}

func (s *testClusteredSuite) TestClusteredHint(c *C) {
	tk := s.newTK(c)
	tk.MustExec("drop table if exists ht")
	tk.MustExec("create table ht (a varchar(64) primary key, b int)")
	tk.MustQuery("select * from ht use index (`PRIMARY`)")
}
