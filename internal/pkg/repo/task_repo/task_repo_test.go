package task_repo_test

import (
	"errors"
	"face-track/internal/pkg/repo/task_repo"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func Test_TaskRepo_CreateTask(t *testing.T) {

	tests := []struct {
		name       string
		beforeTest func(sqlmock.Sqlmock)
		want       int
		wantErr    bool
	}{
		{
			name: "fail create task",
			beforeTest: func(mockSQL sqlmock.Sqlmock) {
				mockSQL.
					ExpectQuery(regexp.QuoteMeta(`
						INSERT INTO task 
							(
							task_status, 
							faces_total, 
							faces_female, 
							faces_male, 
							age_female_avg, 
							age_male_avg
							) 
						VALUES ('new', 0, 0, 0, 0, 0) 
						RETURNING id`,
					)).WithoutArgs().
					WillReturnError(errors.New("whoops, error"))
			},
			wantErr: true,
		},
		{
			name: "success create user",
			beforeTest: func(mockSQL sqlmock.Sqlmock) {
				mockSQL.
					ExpectQuery(regexp.QuoteMeta(`
						INSERT INTO task 
							(
							task_status, 
							faces_total, 
							faces_female, 
							faces_male, 
							age_female_avg, 
							age_male_avg
							) 
						VALUES ('new', 0, 0, 0, 0, 0) 
						RETURNING id`,
					)).WithoutArgs().
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mockSQL, _ := sqlmock.New()
			defer mockDB.Close()

			db := sqlx.NewDb(mockDB, "sqlmock")

			r := task_repo.New(db)

			if tt.beforeTest != nil {
				tt.beforeTest(mockSQL)
			}

			got, err := r.CreateTask()
			if (err != nil) != tt.wantErr {
				t.Errorf("taskRepo.CreateTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("taskRepo.CreateTask() = %v, want %v", got, tt.want)
			}
		})
	}
}
