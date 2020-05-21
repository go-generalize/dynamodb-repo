// +build internal

package tests

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	task "github.com/go-generalize/dynamodb-repo/testfiles/a"
	"github.com/guregu/dynamo"
)

func createTable(t *testing.T, tableName, schema string) {
	t.Helper()

	env := append(
		os.Environ(),
		"AWS_DEFAULT_REGION=ap-northeast-1",
		"AWS_ACCESS_KEY_ID=access_key",
		"AWS_SECRET_ACCESS_KEY=access_secret",
	)

	cmd := exec.Command(
		"aws",
		"dynamodb",
		"delete-table",
		"--endpoint-url",
		"http://localhost:8000",
		"--table-name",
		tableName,
	)
	cmd.Env = env

	b, _ := cmd.CombinedOutput()

	t.Log("delete table", string(b))

	fp, err := ioutil.TempFile("", "*.json")

	if err != nil {
		t.Fatalf("failed to create json: %+v", err)
	}
	io.Copy(
		fp,
		strings.NewReader(schema),
	)
	fp.Close()

	t.Logf("file name: %+v", fp.Name())

	cmd = exec.Command(
		"aws",
		"dynamodb",
		"create-table",
		"--endpoint-url",
		"http://localhost:8000",
		"--cli-input-json",
		"file://"+fp.Name(),
	)
	cmd.Env = env

	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("failed to create table for DynamoDB(%s): %+v", string(output), err)
	}

}

func initDynamoClient(t *testing.T) *dynamo.DB {
	t.Helper()

	if os.Getenv("DYNAMODB_LOCAL_ENDPOINT") == "" {
		os.Setenv("DYNAMODB_LOCAL_ENDPOINT", "http://localhost:8000")
	}

	ep := os.Getenv("DYNAMODB_LOCAL_ENDPOINT")

	client := dynamo.New(session.New(), &aws.Config{
		Region:     aws.String("ap-northeast-1"),
		Endpoint:   aws.String(ep),
		DisableSSL: aws.Bool(true),
	})

	createTable(t, "Name", task.NameSchema)
	createTable(t, "Task", task.TaskSchema)

	return client
}

func compareTask(t *testing.T, expected, actual *task.Task) {
	t.Helper()

	if actual.ID != expected.ID {
		t.Fatalf("unexpected id: %d(expected: %d)", actual.ID, expected.ID)
	}

	if !actual.Created.Equal(expected.Created) {
		t.Fatalf("unexpected time: %s(expected: %s)", actual.Created, expected.Created)
	}

	if actual.Desc != expected.Desc {
		t.Fatalf("unexpected desc: %s(expected: %s)", actual.Desc, expected.Created)
	}

	if actual.Done != expected.Done {
		t.Fatalf("unexpected done: %v(expected: %v)", actual.Done, expected.Done)
	}
}

func TestDatastoreTransactionTask(t *testing.T) {
	client := initDynamoClient(t)

	taskRepo := task.NewTaskRepository(client)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	var ids []int64
	defer func() {
		defer cancel()
		if err := taskRepo.DeleteMultiByIDs(ctx, ids); err != nil {
			t.Fatal(err)
		}
	}()

	now := time.Unix(0, time.Now().UnixNano())
	desc := "Hello, World!"

	t.Run("Multi", func(tr *testing.T) {
		tks := make([]*task.Task, 0)
		for i := int64(1); i <= 10; i++ {
			tk := &task.Task{
				ID:         i * 100,
				Desc:       fmt.Sprintf("%s%d", desc, i),
				Created:    now,
				Done:       true,
				Done2:      false,
				Count:      int(i),
				Count64:    0,
				Proportion: 0.12345 + float64(i),
				Flag:       task.Flag(true),
				NameList:   []string{"a", "b", "c"},
			}
			tks = append(tks, tk)
			ids = append(ids, tk.ID)
		}
		err := taskRepo.InsertMulti(ctx, tks)
		if err != nil {
			tr.Fatalf("%+v", err)
		}

		tks2 := make([]*task.Task, 0)
		for i := int64(1); i <= 10; i++ {
			tk := &task.Task{
				ID:         i * 100,
				Desc:       fmt.Sprintf("%s%d", desc, i),
				Created:    now,
				Done:       true,
				Done2:      false,
				Count:      int(i),
				Count64:    i,
				Proportion: 0.12345 + float64(i),
				Flag:       task.Flag(true),
				NameList:   []string{"a", "b", "c"},
			}
			tks2 = append(tks2, tk)
		}
		if err := taskRepo.UpdateMulti(ctx, tks2); err != nil {
			tr.Fatalf("%+v", err)
		}

		if tks[0].ID != tks2[0].ID {
			tr.Fatalf("unexpected id: %d (expected: %d)", tks[0].ID, tks2[0].ID)
		}
	})

	t.Run("Single", func(tr *testing.T) {
		tk := &task.Task{
			ID:         1001,
			Desc:       fmt.Sprintf("%s%d", desc, 1001),
			Created:    now,
			Done:       true,
			Done2:      false,
			Count:      11,
			Count64:    11,
			Proportion: 0.12345 + 11,
			NameList:   []string{"a", "b", "c"},
		}
		err := taskRepo.Insert(ctx, tk)
		if err != nil {
			tr.Fatalf("%+v", err)
		}
		ids = append(ids, tk.ID)

		tk.Count = 12
		if err := taskRepo.Update(ctx, tk); err != nil {
			tr.Fatalf("%+v", err)
		}

		tsk, err := taskRepo.Get(ctx, tk.ID)
		if err != nil {
			tr.Fatalf("%+v", err)
		}

		if tsk.Count != 12 {
			tr.Fatalf("unexpected Count: %d (expected: %d)", tsk.Count, 12)
		}
	})
}

func TestDatastoreListTask(t *testing.T) {
	client := initDynamoClient(t)

	taskRepo := task.NewTaskRepository(client)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	var ids []int64
	defer func() {
		defer cancel()
		if err := taskRepo.DeleteMultiByIDs(ctx, ids); err != nil {
			t.Fatal(err)
		}
	}()

	now := time.Unix(0, time.Now().UnixNano())
	desc := "Hello, World!"

	tks := make([]*task.Task, 0)
	for i := int64(1); i <= 10; i++ {
		tk := &task.Task{
			ID:         i * 100,
			Desc:       fmt.Sprintf("%s%d", desc, i),
			Created:    now,
			Done:       true,
			Done2:      false,
			Count:      int(i),
			Count64:    0,
			Proportion: 0.12345 + float64(i),
			Flag:       task.Flag(true),
			NameList:   []string{"a", "b", "c"},
		}
		tks = append(tks, tk)
		ids = append(ids, tk.ID)
	}
	err := taskRepo.InsertMulti(ctx, tks)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	t.Run("int(1件)", func(t *testing.T) {
		var tasks []*task.Task

		err := taskRepo.List("count", 1).AllWithContext(ctx, &tasks)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if len(tasks) != 1 {
			t.Fatal("not match")
		}
	})

	t.Run("float(1件)", func(t *testing.T) {
		var tasks []*task.Task

		prop := task.NumericCriteriaBase.Parse(1.12345)
		err := taskRepo.List("proportion", prop).AllWithContext(ctx, &tasks)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if len(tasks) != 1 {
			t.Fatal("not match")
		}
	})

	t.Run("bool(10件)", func(t *testing.T) {
		var tasks []*task.Task

		err := taskRepo.List("done", true).AllWithContext(ctx, &tasks)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if len(tasks) != 10 {
			t.Fatal("not match")
		}
	})

	t.Run("time.Time(10件)", func(t *testing.T) {
		var tasks []*task.Task

		err := taskRepo.List("created", now).AllWithContext(ctx, &tasks)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if len(tasks) != 10 {
			t.Fatal("not match")
		}
	})
}

func TestDatastoreListNameWithIndexes(t *testing.T) {
	client := initDynamoClient(t)

	nameRepo := task.NewNameRepository(client)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	var ids []int64
	defer func() {
		defer cancel()
		if err := nameRepo.DeleteMultiByIDs(ctx, ids); err != nil {
			t.Fatal(err)
		}
	}()

	now := time.Unix(0, time.Now().UnixNano())
	desc := "Hello, World!"
	desc2 := "Prefix, Test!"

	tks := make([]*task.Name, 0)
	for i := int64(1); i <= 10; i++ {
		tk := &task.Name{
			ID:        i,
			Created:   now,
			Desc:      fmt.Sprintf("%s%d", desc, i),
			Desc2:     fmt.Sprintf("%s%d", desc2, i),
			Done:      true,
			Count:     int(i),
			PriceList: []int{1, 2, 3, 4, 5},
		}
		tks = append(tks, tk)
		ids = append(ids, i)
	}

	err := nameRepo.InsertMulti(ctx, tks)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	t.Run("int(1件)", func(t *testing.T) {
		var tasks []*task.Name

		err := nameRepo.List("count", 1).AllWithContext(ctx, &tasks)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if len(tasks) != 1 {
			t.Fatal("not match")
		}
	})

	t.Run("bool(10件)", func(t *testing.T) {
		var tasks []*task.Name

		err := nameRepo.List("done", true).AllWithContext(ctx, &tasks)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if len(tasks) != 10 {
			t.Fatal("not match")
		}
	})

	t.Run("like(10件)", func(t *testing.T) {
		var tasks []*task.Name

		err := nameRepo.List("description", "ll").AllWithContext(ctx, &tasks)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if len(tasks) != 10 {
			t.Fatal("not match")
		}
	})

	t.Run("prefix", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			var tasks []*task.Name

			err := nameRepo.List("description2", "Pre").AllWithContext(ctx, &tasks)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if len(tasks) != 10 {
				t.Fatal("not match")
			}
		})

		t.Run("Failure", func(t *testing.T) {
			var tasks []*task.Name

			err := nameRepo.List("description2", "He").AllWithContext(ctx, &tasks)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if len(tasks) != 0 {
				t.Fatal("not match")
			}
		})
	})

	t.Run("time.Time(10件)", func(t *testing.T) {
		var tasks []*task.Name

		err := nameRepo.List("created", now).AllWithContext(ctx, &tasks)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if len(tasks) != 10 {
			t.Fatal("not match")
		}
	})
}

func TestDatastore(t *testing.T) {
	client := initDynamoClient(t)

	taskRepo := task.NewTaskRepository(client)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	now := time.Unix(time.Now().Unix(), 0)
	desc := "hello"

	id := int64(1234)
	err := taskRepo.Insert(ctx, &task.Task{
		ID:      id,
		Desc:    desc,
		Created: now,
		Done:    true,
	})

	if err != nil {
		t.Fatalf("failed to put item: %+v", err)
	}

	ret, err := taskRepo.Get(ctx, id)

	if err != nil {
		t.Fatalf("failed to get item: %+v", err)
	}

	compareTask(t, &task.Task{
		ID:      id,
		Desc:    desc,
		Created: now,
		Done:    true,
	}, ret)

	rets, err := taskRepo.GetMulti(ctx, []int64{id})

	if err != nil {
		t.Fatalf("failed to get item: %+v", err)
	}

	if len(rets) != 1 {
		t.Errorf("GetMulti should return 1 item: %+v", err)
	}

	compareTask(t, &task.Task{
		ID:      id,
		Desc:    desc,
		Created: now,
		Done:    true,
	}, rets[0])

	compareTask(t, &task.Task{
		ID:      id,
		Desc:    desc,
		Created: now,
		Done:    true,
	}, ret)

	if err := taskRepo.DeleteByID(ctx, id); err != nil {
		t.Fatalf("delete failed: %+v", err)
	}

	if _, err := taskRepo.Get(ctx, id); err != datastore.ErrNoSuchEntity {
		t.Fatalf("Get deleted item should return ErrNoSuchEntity: %+v", err)
	}
}
