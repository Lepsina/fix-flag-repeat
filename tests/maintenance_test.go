package tests

import (
	"fmt"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_Maintenance"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Auth"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Discovery"
	"google.golang.org/protobuf/proto"

	"github.com/ydb-platform/ydbops/pkg/client/cms"
	"github.com/ydb-platform/ydbops/tests/mock"
)

var _ = Describe("Test Maintenance", func() {
	BeforeEach(RunBeforeEach)
	AfterEach(RunAfterEach)

	DescribeTable("maintenance", RunTestCase,
		Entry("restart two storage hosts by specifying FQDN, storage-only baremetal cluster", TestCase{
			nodeConfiguration: [][]uint32{
				{1, 2, 3, 4, 5, 6, 7, 8},
			},
			nodeInfoMap: map[uint32]mock.TestNodeInfo{},
			steps: []StepData{
				{
					ydbopsInvocation: Command{
						"--endpoint", "grpcs://localhost:2135",
						"--verbose",
						"--user", mock.TestUser,
						"--ca-file", filepath.Join(".", "test-data", "ssl-data", "ca.crt"),
						"maintenance",
						"create",
						"--duration", "180",
						"--availability-mode", "strong",
						"--hosts=ydb-1.ydb.tech,ydb-2.ydb.tech",
					},
					expectedRequests: []proto.Message{
						&Ydb_Auth.LoginRequest{
							User:     mock.TestUser,
							Password: mock.TestPassword,
						},
						&Ydb_Maintenance.ListClusterNodesRequest{},
						&Ydb_Maintenance.CreateMaintenanceTaskRequest{
							TaskOptions: &Ydb_Maintenance.MaintenanceTaskOptions{
								TaskUid:          "task-uuid-1",
								Description:      "Rolling restart maintenance task",
								AvailabilityMode: Ydb_Maintenance.AvailabilityMode_AVAILABILITY_MODE_STRONG,
							},
							ActionGroups: mock.MakeActionGroupsFromHostFQDNsFixedDuration(time.Second * 180, "ydb-1.ydb.tech", "ydb-2.ydb.tech"),
						},
					},
					expectedOutputRegexps: []string{
						// Your task id is:\n\n<uuid>\n\nPlease write it down for refreshing and completing the task later.\n
						fmt.Sprintf("Your task id is:\n\n%s%s\n\n", cms.TaskUuidPrefix, uuidRegexpString),
					},
				},
				{
					ydbopsInvocation: Command{
						"--endpoint", "grpcs://localhost:2135",
						"--verbose",
						"--user", mock.TestUser,
						"--ca-file", filepath.Join(".", "test-data", "ssl-data", "ca.crt"),
						"maintenance",
						"list",
					},
					expectedRequests: []proto.Message{
						&Ydb_Auth.LoginRequest{
							User:     mock.TestUser,
							Password: mock.TestPassword,
						},
						&Ydb_Discovery.WhoAmIRequest{},
						&Ydb_Maintenance.ListMaintenanceTasksRequest{
							User: &mock.TestUser,
						},
						&Ydb_Maintenance.GetMaintenanceTaskRequest{
							TaskUid: "task-uuid-1",
						},
					},
					expectedOutputRegexps: []string{
						fmt.Sprintf("Uid: %s%s\n", cms.TaskUuidPrefix, uuidRegexpString),
						"  Lock on host ydb-1.ydb.tech",
						"PERFORMED",
						"  Lock on host ydb-2.ydb.tech",
						"PENDING, (\\S+)",
					},
				},
				{
					ydbopsInvocation: Command{
						"--endpoint", "grpcs://localhost:2135",
						"--verbose",
						"--user", mock.TestUser,
						"--ca-file", filepath.Join(".", "test-data", "ssl-data", "ca.crt"),
						"maintenance",
						"complete",
						"--task-id",
						testWillInsertTaskUuid,
						"--hosts=ydb-1.ydb.tech",
					},
					expectedRequests: []proto.Message{
						&Ydb_Auth.LoginRequest{
							User:     mock.TestUser,
							Password: mock.TestPassword,
						},
						&Ydb_Maintenance.GetMaintenanceTaskRequest{
							TaskUid: "task-UUID-1",
						},
						&Ydb_Maintenance.CompleteActionRequest{
							ActionUids: []*Ydb_Maintenance.ActionUid{
								{
									TaskUid:  "task-UUID-1",
									GroupId:  "group-UUID-1",
									ActionId: "action-UUID-1",
								},
							},
						},
					},
					expectedOutputRegexps: []string{
						fmt.Sprintf("  Completed action id: %s, status: SUCCESS", uuidRegexpString),
					},
				},
				{
					ydbopsInvocation: Command{
						"--endpoint", "grpcs://localhost:2135",
						"--verbose",
						"--user", mock.TestUser,
						"--ca-file", filepath.Join(".", "test-data", "ssl-data", "ca.crt"),
						"maintenance",
						"refresh",
						"--task-id",
						testWillInsertTaskUuid,
					},
					expectedRequests: []proto.Message{
						&Ydb_Auth.LoginRequest{
							User:     mock.TestUser,
							Password: mock.TestPassword,
						},
						&Ydb_Maintenance.RefreshMaintenanceTaskRequest{
							TaskUid: "task-uuid-1",
						},
					},
					expectedOutputRegexps: []string{
						fmt.Sprintf("Uid: %s%s\n", cms.TaskUuidPrefix, uuidRegexpString),
						"  Lock on host ydb-2.ydb.tech",
						"PERFORMED",
					},
				},
				{
					ydbopsInvocation: Command{
						"--endpoint", "grpcs://localhost:2135",
						"--verbose",
						"--user", mock.TestUser,
						"--ca-file", filepath.Join(".", "test-data", "ssl-data", "ca.crt"),
						"maintenance",
						"complete",
						"--task-id",
						testWillInsertTaskUuid,
						"--hosts=ydb-2.ydb.tech",
					},
					expectedRequests: []proto.Message{
						&Ydb_Auth.LoginRequest{
							User:     mock.TestUser,
							Password: mock.TestPassword,
						},
						&Ydb_Maintenance.GetMaintenanceTaskRequest{
							TaskUid: "task-UUID-1",
						},
						&Ydb_Maintenance.CompleteActionRequest{
							ActionUids: []*Ydb_Maintenance.ActionUid{
								{
									TaskUid:  "task-UUID-1",
									GroupId:  "group-UUID-1",
									ActionId: "action-UUID-2",
								},
							},
						},
					},
					expectedOutputRegexps: []string{
						fmt.Sprintf("  Completed action id: %s, status: SUCCESS", uuidRegexpString),
					},
				},
				{
					ydbopsInvocation: Command{
						"--endpoint", "grpcs://localhost:2135",
						"--verbose",
						"--user", mock.TestUser,
						"--ca-file", filepath.Join(".", "test-data", "ssl-data", "ca.crt"),
						"maintenance",
						"list",
					},
					expectedRequests: []proto.Message{
						&Ydb_Auth.LoginRequest{
							User:     mock.TestUser,
							Password: mock.TestPassword,
						},
						&Ydb_Discovery.WhoAmIRequest{},
						&Ydb_Maintenance.ListMaintenanceTasksRequest{
							User: &mock.TestUser,
						},
					},
					expectedOutputRegexps: []string{
						"There are no maintenance tasks",
					},
				},
			},
		},
		),
		Entry("restart two storage hosts by specifying node ids, storage-only baremetal cluster", TestCase{
			nodeConfiguration: [][]uint32{
				{1, 2, 3, 4, 5, 6, 7, 8},
			},
			nodeInfoMap: map[uint32]mock.TestNodeInfo{},
			steps: []StepData{
				{
					ydbopsInvocation: Command{
						"--endpoint", "grpcs://localhost:2135",
						"--verbose",
						"--user", mock.TestUser,
						"--ca-file", filepath.Join(".", "test-data", "ssl-data", "ca.crt"),
						"maintenance",
						"create",
						"--duration", "180",
						"--availability-mode", "strong",
						"--hosts=1,2",
					},
					expectedRequests: []proto.Message{
						&Ydb_Auth.LoginRequest{
							User:     mock.TestUser,
							Password: mock.TestPassword,
						},
						&Ydb_Maintenance.ListClusterNodesRequest{},
						&Ydb_Maintenance.CreateMaintenanceTaskRequest{
							TaskOptions: &Ydb_Maintenance.MaintenanceTaskOptions{
								TaskUid:          "task-uuid-1",
								Description:      "Rolling restart maintenance task",
								AvailabilityMode: Ydb_Maintenance.AvailabilityMode_AVAILABILITY_MODE_STRONG,
							},
							ActionGroups: mock.MakeActionGroupsFromNodesIdsFixedDuration(time.Second * 180, 1, 2),
						},
					},
					expectedOutputRegexps: []string{
						// Your task id is:\n\n<uuid>\n\nPlease write it down for refreshing and completing the task later.\n
						fmt.Sprintf("Your task id is:\n\n%s%s\n\n", cms.TaskUuidPrefix, uuidRegexpString),
					},
				},
				{
					ydbopsInvocation: Command{
						"--endpoint", "grpcs://localhost:2135",
						"--verbose",
						"--user", mock.TestUser,
						"--ca-file", filepath.Join(".", "test-data", "ssl-data", "ca.crt"),
						"maintenance",
						"list",
					},
					expectedRequests: []proto.Message{
						&Ydb_Auth.LoginRequest{
							User:     mock.TestUser,
							Password: mock.TestPassword,
						},
						&Ydb_Discovery.WhoAmIRequest{},
						&Ydb_Maintenance.ListMaintenanceTasksRequest{
							User: &mock.TestUser,
						},
						&Ydb_Maintenance.GetMaintenanceTaskRequest{
							TaskUid: "task-uuid-1",
						},
					},
					expectedOutputRegexps: []string{
						fmt.Sprintf("Uid: %s%s\n", cms.TaskUuidPrefix, uuidRegexpString),
						"  Lock on node 1",
						"PERFORMED",
						"  Lock on node 2",
						"PENDING, (\\S+)",
					},
				},
				{
					ydbopsInvocation: Command{
						"--endpoint", "grpcs://localhost:2135",
						"--verbose",
						"--user", mock.TestUser,
						"--ca-file", filepath.Join(".", "test-data", "ssl-data", "ca.crt"),
						"maintenance",
						"complete",
						"--task-id",
						testWillInsertTaskUuid,
						"--hosts=1",
					},
					expectedRequests: []proto.Message{
						&Ydb_Auth.LoginRequest{
							User:     mock.TestUser,
							Password: mock.TestPassword,
						},
						&Ydb_Maintenance.GetMaintenanceTaskRequest{
							TaskUid: "task-UUID-1",
						},
						&Ydb_Maintenance.CompleteActionRequest{
							ActionUids: []*Ydb_Maintenance.ActionUid{
								{
									TaskUid:  "task-UUID-1",
									GroupId:  "group-UUID-1",
									ActionId: "action-UUID-1",
								},
							},
						},
					},
					expectedOutputRegexps: []string{
						fmt.Sprintf("  Completed action id: %s, status: SUCCESS", uuidRegexpString),
					},
				},
				{
					ydbopsInvocation: Command{
						"--endpoint", "grpcs://localhost:2135",
						"--verbose",
						"--user", mock.TestUser,
						"--ca-file", filepath.Join(".", "test-data", "ssl-data", "ca.crt"),
						"maintenance",
						"refresh",
						"--task-id",
						testWillInsertTaskUuid,
					},
					expectedRequests: []proto.Message{
						&Ydb_Auth.LoginRequest{
							User:     mock.TestUser,
							Password: mock.TestPassword,
						},
						&Ydb_Maintenance.RefreshMaintenanceTaskRequest{
							TaskUid: "task-uuid-1",
						},
					},
					expectedOutputRegexps: []string{
						fmt.Sprintf("Uid: %s%s\n", cms.TaskUuidPrefix, uuidRegexpString),
						"  Lock on node 2",
						"PERFORMED",
					},
				},
				{
					ydbopsInvocation: Command{
						"--endpoint", "grpcs://localhost:2135",
						"--verbose",
						"--user", mock.TestUser,
						"--ca-file", filepath.Join(".", "test-data", "ssl-data", "ca.crt"),
						"maintenance",
						"complete",
						"--task-id",
						testWillInsertTaskUuid,
						"--hosts=2",
					},
					expectedRequests: []proto.Message{
						&Ydb_Auth.LoginRequest{
							User:     mock.TestUser,
							Password: mock.TestPassword,
						},
						&Ydb_Maintenance.GetMaintenanceTaskRequest{
							TaskUid: "task-UUID-1",
						},
						&Ydb_Maintenance.CompleteActionRequest{
							ActionUids: []*Ydb_Maintenance.ActionUid{
								{
									TaskUid:  "task-UUID-1",
									GroupId:  "group-UUID-1",
									ActionId: "action-UUID-2",
								},
							},
						},
					},
					expectedOutputRegexps: []string{
						fmt.Sprintf("  Completed action id: %s, status: SUCCESS", uuidRegexpString),
					},
				},
				{
					ydbopsInvocation: Command{
						"--endpoint", "grpcs://localhost:2135",
						"--verbose",
						"--user", mock.TestUser,
						"--ca-file", filepath.Join(".", "test-data", "ssl-data", "ca.crt"),
						"maintenance",
						"list",
					},
					expectedRequests: []proto.Message{
						&Ydb_Auth.LoginRequest{
							User:     mock.TestUser,
							Password: mock.TestPassword,
						},
						&Ydb_Discovery.WhoAmIRequest{},
						&Ydb_Maintenance.ListMaintenanceTasksRequest{
							User: &mock.TestUser,
						},
					},
					expectedOutputRegexps: []string{
						"There are no maintenance tasks",
					},
				},
			},
		},
		),
	)
})
