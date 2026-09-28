// Copyright 2019 Yunion
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

package qcloud

import (
	"testing"
	"time"

	billingapi "yunion.io/x/cloudmux/pkg/apis/billing"
	api "yunion.io/x/cloudmux/pkg/apis/compute"
	"yunion.io/x/cloudmux/pkg/cloudprovider"
	"yunion.io/x/jsonutils"
)

const postgresSampleJSON = `{
  "Region": "ap-guangzhou",
  "Zone": "ap-guangzhou-3",
  "ProjectId": 0,
  "VpcId": "vpc-abc123",
  "SubnetId": "subnet-abc123",
  "DBInstanceId": "postgres-6r233v55",
  "DBInstanceName": "pg-test",
  "DBInstanceStatus": "running",
  "DBInstanceMemory": 4,
  "DBInstanceStorage": 100,
  "DBInstanceCpu": 2,
  "DBInstanceClass": "cdb.pg.z1.4g",
  "DBMajorVersion": "13",
  "DBVersion": "13.3",
  "DBInstanceType": "primary",
  "DBInstanceVersion": "standard",
  "CreateTime": "2025-09-16 20:27:09",
  "ExpireTime": "0000-00-00 00:00:00",
  "PayType": "postpaid",
  "AutoRenew": 0,
  "DBInstanceNetInfo": [
    {"Address": "", "Ip": "10.0.0.8", "Port": 5432, "NetType": "private", "Status": "opened", "VpcId": "vpc-abc123", "SubnetId": "subnet-abc123"},
    {"Address": "postgres-6r233v55.sql.tencentcdb.com", "Ip": "", "Port": 21000, "NetType": "public", "Status": "opened"}
  ],
  "DBNodeSet": [
    {"Role": "Primary", "Zone": "ap-guangzhou-3"},
    {"Role": "Standby", "Zone": "ap-guangzhou-4"}
  ],
  "TagList": [{"TagKey": "env", "TagValue": "test"}]
}`

func loadPostgresSample(t *testing.T) *SPostgreSQL {
	t.Helper()
	obj, err := jsonutils.ParseString(postgresSampleJSON)
	if err != nil {
		t.Fatalf("parse sample: %v", err)
	}
	pg := &SPostgreSQL{}
	if err := obj.Unmarshal(pg); err != nil {
		t.Fatalf("unmarshal sample: %v", err)
	}
	return pg
}

func TestPostgreSQLImplementsInterface(t *testing.T) {
	var _ cloudprovider.ICloudDBInstance = &SPostgreSQL{}
}

func TestPostgreSQLFields(t *testing.T) {
	pg := loadPostgresSample(t)

	checks := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"id", pg.GetGlobalId(), "postgres-6r233v55"},
		{"name", pg.GetName(), "pg-test"},
		{"status", pg.GetStatus(), api.DBINSTANCE_RUNNING},
		{"engine", pg.GetEngine(), api.DBINSTANCE_TYPE_POSTGRESQL},
		{"engineVersion", pg.GetEngineVersion(), "13"},
		{"instanceType", pg.GetInstanceType(), "cdb.pg.z1.4g"},
		{"cpu", pg.GetVcpuCount(), 2},
		{"memMb", pg.GetVmemSizeMB(), 4096},
		{"diskGb", pg.GetDiskSizeGB(), 100},
		{"category", pg.GetCategory(), api.QCLOUD_DBINSTANCE_CATEGORY_HA},
		{"vpc", pg.GetIVpcId(), "vpc-abc123"},
		{"port", pg.GetPort(), 5432},
		{"internalConn", pg.GetInternalConnectionStr(), "10.0.0.8:5432"},
		{"publicConn", pg.GetConnectionStr(), "postgres-6r233v55.sql.tencentcdb.com:21000"},
		{"zone1", pg.GetZone1Id(), "ap-guangzhou-3"},
		{"zone2", pg.GetZone2Id(), "ap-guangzhou-4"},
		{"master", pg.GetMasterInstanceId(), ""},
		{"projectId", pg.GetProjectId(), "0"},
		{"billing", pg.GetBillingType(), billingapi.BILLING_TYPE_POSTPAID},
		{"expired", pg.GetExpiredAt().IsZero(), true},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, c.got, c.want)
		}
	}

	wantCreated := time.Date(2025, 9, 16, 12, 27, 9, 0, time.UTC)
	if !pg.GetCreatedAt().Equal(wantCreated) {
		t.Errorf("createdAt: got %v, want %v", pg.GetCreatedAt(), wantCreated)
	}

	nets, err := pg.GetDBNetworks()
	if err != nil || len(nets) != 1 || nets[0].IP != "10.0.0.8" || nets[0].NetworkId != "subnet-abc123" {
		t.Errorf("networks: got %+v, err %v", nets, err)
	}

	tags, _ := pg.GetTags()
	if tags["env"] != "test" {
		t.Errorf("tags: got %v", tags)
	}
}

func TestPostgreSQLReadonlyAndStatus(t *testing.T) {
	pg := &SPostgreSQL{
		DBInstanceType:     "readonly",
		MasterDBInstanceId: "postgres-master",
		DBNodeSet:          []SPostgresNode{{Role: "Primary", Zone: "z1"}},
		DBInstanceStatus:   "someNewStatus",
		PayType:            "prepaid",
		ExpireTime:         "2026-01-01 08:00:00",
	}
	if pg.GetMasterInstanceId() != "postgres-master" {
		t.Errorf("master: got %q", pg.GetMasterInstanceId())
	}
	if pg.GetCategory() != api.QCLOUD_DBINSTANCE_CATEGORY_BASIC {
		t.Errorf("category: got %q", pg.GetCategory())
	}
	if pg.GetStatus() != api.DBINSTANCE_UNKNOWN {
		t.Errorf("status: got %q", pg.GetStatus())
	}
	if pg.GetBillingType() != billingapi.BILLING_TYPE_PREPAID {
		t.Errorf("billing: got %q", pg.GetBillingType())
	}
	want := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if !pg.GetExpiredAt().Equal(want) {
		t.Errorf("expired: got %v, want %v", pg.GetExpiredAt(), want)
	}
	if pg.GetConnectionStr() != "" || pg.GetInternalConnectionStr() != "" || pg.GetPort() != 0 {
		t.Errorf("empty net info should yield empty connection")
	}
}

// 腾讯云 SDK 响应经 encoding/json 解析后数字为 float64，ProjectId 不能被格式化成 "0.000000"
func TestPostgreSQLProjectIdFromFloat(t *testing.T) {
	for _, projectId := range []float64{0, 1234567} {
		obj := jsonutils.Marshal(map[string]interface{}{"DBInstanceId": "postgres-x", "ProjectId": projectId})
		pg := &SPostgreSQL{}
		if err := obj.Unmarshal(pg); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		want := jsonutils.NewInt(int64(projectId)).String()
		if pg.GetProjectId() != want {
			t.Errorf("projectId: got %q, want %q", pg.GetProjectId(), want)
		}
	}
}

func TestParsePostgresTime(t *testing.T) {
	for _, s := range []string{"", "0000-00-00 00:00:00", "invalid"} {
		if !parsePostgresTime(s).IsZero() {
			t.Errorf("parsePostgresTime(%q) should be zero", s)
		}
	}
}
