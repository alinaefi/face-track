package task_repo_test

import (
	"database/sql"
	"errors"
	"face-track/internal/pkg/model/task_model"
	"face-track/internal/pkg/repo/task_repo"
	"face-track/tools"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func Test_TaskRepo_CreateTask(t *testing.T) {

	// Define table-driven tests
	tests := []struct {
		name       string                // Name of the test case
		beforeTest func(sqlmock.Sqlmock) // Setup mock expectations
		want       int                   // Expected task ID
		wantErr    bool                  // Whether we expect an error
	}{
		{
			name: "fail create task",
			beforeTest: func(mockSQL sqlmock.Sqlmock) {
				// Simulate a DB error during INSERT
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
					WillReturnError(errors.New("whoops, error")) // Mock DB failure
			},
			wantErr: true, // We expect an error here
		},
		{
			name: "success create task",
			beforeTest: func(mockSQL sqlmock.Sqlmock) {
				// Simulate a successful DB insert that returns id = 1
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
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1)) // Simulate return row
			},
			want: 1, // We expect the returned task ID to be 1
		},
	}

	// Run all test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock DB and SQLMock
			mockDB, mockSQL, _ := sqlmock.New()
			defer mockDB.Close()

			// Wrap sqlmock.DB with sqlx for our repository
			db := sqlx.NewDb(mockDB, "sqlmock")

			// Initialize the repo with the mocked DB
			r := task_repo.New(db)

			// Set up test-specific expectations
			if tt.beforeTest != nil {
				tt.beforeTest(mockSQL)
			}

			// Call the function under test
			got, err := r.CreateTask()

			// Check if the error matches expected outcome
			if (err != nil) != tt.wantErr {
				t.Errorf("taskRepo.CreateTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check if the returned task ID matches expectation
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("taskRepo.CreateTask() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_TaskRepo_GetTaskById(t *testing.T) {

	type args struct {
		taskId int
	}

	// Define table-driven tests
	tests := []struct {
		name          string
		args          args
		beforeTest    func(sqlmock.Sqlmock)
		want          *task_model.Task
		wantErr       bool
		wantErrorType error
	}{
		{ // task with given id not found
			name: "fail retrieve task: task not found",
			args: args{taskId: 1},
			beforeTest: func(mockSQL sqlmock.Sqlmock) {
				mockSQL.
					ExpectQuery(regexp.QuoteMeta(
						`SELECT 
							id, 
							task_status, 
							faces_total, 
							faces_female, 
							faces_male, 
							age_female_avg, 
							age_male_avg 
						FROM task 
						WHERE id=$1`,
					)).WithArgs(1).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr:       true,
			wantErrorType: tools.ErrNotFound,
		},

		{ // failed retrieve task
			name: "fail retrieve task",
			args: args{taskId: 1},
			beforeTest: func(mockSQL sqlmock.Sqlmock) {
				mockSQL.
					ExpectQuery(regexp.QuoteMeta(
						`SELECT 
							id, 
							task_status, 
							faces_total, 
							faces_female, 
							faces_male, 
							age_female_avg, 
							age_male_avg 
						FROM task 
						WHERE id=$1`,
					)).WithArgs(1).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},

		{ // successfully retrieved task
			name: "success retrieve task",
			args: args{taskId: 1},
			beforeTest: func(mockSQL sqlmock.Sqlmock) {
				mockSQL.
					ExpectQuery(regexp.QuoteMeta(
						`SELECT 
							id, 
							task_status, 
							faces_total, 
							faces_female, 
							faces_male, 
							age_female_avg, 
							age_male_avg 
						FROM task 
						WHERE id=$1`,
					)).WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "task_status", "faces_total", "faces_female", "faces_male", "age_female_avg", "age_male_avg"}).AddRow(1, "", 0, 0, 0, 0, 0))
			},
			want:    &task_model.Task{Id: 1},
			wantErr: false,
		},
	}

	// Run all test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mockSQL, _ := sqlmock.New()
			defer mockDB.Close()

			db := sqlx.NewDb(mockDB, "sqlmock")

			r := task_repo.New(db)

			if tt.beforeTest != nil {
				tt.beforeTest(mockSQL)
			}

			got, err := r.GetTaskById(tt.args.taskId)

			if (err != nil) != tt.wantErr {
				t.Errorf("taskRepo.GetTaskById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check if the error matches the expected error type
			if tt.wantErrorType != nil && !errors.Is(err, tt.wantErrorType) {
				t.Errorf("expected error type %v, got %v", tt.wantErrorType, err)
			}

			// Check if the returned result matches the expected value
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("taskRepo.GetTaskById() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_TaskRepo_GetTaskImages(t *testing.T) {

	// define arguments
	type args struct {
		taskId int
	}

	// define tests
	tests := []struct {
		name       string
		args       args
		beforeTest func(sqlmock.Sqlmock)
		want       []*task_model.Image
		wantErr    bool
	}{
		{ // fail retrieve images
			name: "fail retrieve images",
			args: args{taskId: 1},
			beforeTest: func(mockSQL sqlmock.Sqlmock) {
				mockSQL.
					ExpectQuery(regexp.QuoteMeta(
						`SELECT 
							id, 
							task_id, 
							image_name, 
							done 
						FROM task_image 
						WHERE task_id=$1`,
					)).WithArgs(1).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},

		{ // success retrieve images
			name: "success retrieving task images",
			args: args{taskId: 1},
			beforeTest: func(mockSQL sqlmock.Sqlmock) {
				mockSQL.
					ExpectQuery(regexp.QuoteMeta(
						`SELECT 
							id, 
							task_id, 
							image_name, 
							done 
						FROM task_image 
						WHERE task_id=$1`,
					)).WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "task_id", "image_name", "done"}).AddRow(2, 1, "", false))
			},
			want:    []*task_model.Image{{Id: 2, TaskId: 1}},
			wantErr: false,
		},
	}

	// run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mockSQL, _ := sqlmock.New()
			defer mockDB.Close()

			db := sqlx.NewDb(mockDB, "sqlmock")

			r := task_repo.New(db)

			if tt.beforeTest != nil {
				tt.beforeTest(mockSQL)
			}

			got, err := r.GetTaskImages(tt.args.taskId)

			if (err != nil) != tt.wantErr {
				t.Errorf("taskRepo.GetTaskImages() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("taskRepo.GetTaskImages() = %v, want = %v", got, tt.want)
			}
		})
	}
}

func Test_TaskRepo_GetFacesByImageIds(t *testing.T) {

	type args struct {
		imageIds []int
	}

	tests := []struct {
		name       string
		args       args
		beforeTest func(sqlmock.Sqlmock)
		want       map[int][]*task_model.Face
		wantErr    bool
	}{
		{ // no image ids provided
			name: "fail retrieve faces: no image ids provided",
			args: args{imageIds: []int{}},
			beforeTest: func(mockSQL sqlmock.Sqlmock) {
				mockSQL.ExpectQuery(regexp.QuoteMeta(
					`SELECT 
						id, 
						image_id, 
						gender, 
						age, 
						bbox_height, 
						bbox_width, 
						bbox_x, 
						bbox_y 
					FROM face 
					WHERE image_id IN (?)`,
				)).WithArgs([]int{}).
					WillReturnError(errors.New("sql error"))
			},
			wantErr: true,
		},
		{ // success retrieving faces
			name: "success retrieving faces",
			args: args{imageIds: []int{3, 4, 5}},
			beforeTest: func(mockSQL sqlmock.Sqlmock) {
				mockSQL.ExpectQuery(regexp.QuoteMeta(
					`SELECT 
						id, 
						image_id, 
						gender, 
						age, 
						bbox_height, 
						bbox_width, 
						bbox_x, 
						bbox_y 
					FROM face 
					WHERE image_id IN (?, ?, ?)`,
				)).WithArgs(3, 4, 5).
					WillReturnRows(sqlmock.NewRows([]string{"id", "image_id", "gender", "age", "bbox_height", "bbox_width", "bbox_x", "bbox_y"}).AddRow(2, 3, "male", 34, 700, 600, 1088, 904))
			},
			want:    map[int][]*task_model.Face{3: {&task_model.Face{Id: 2, ImageId: 3}}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockDB, mockSQL, _ := sqlmock.New()

			db := sqlx.NewDb(mockDB, "sqlmock")

			r := task_repo.New(db)

			if tt.beforeTest != nil {
				tt.beforeTest(mockSQL)
			}

			// Act
			got, err := r.GetFacesByImageIds(tt.args.imageIds)

			// Assert
			if (err != nil) != tt.wantErr {
				t.Errorf("taskRepo.GetFacesByImageIds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Local helper to compare only Id and ImageId fields
			equalFacesByIdAndImageId := func(a, b map[int][]*task_model.Face) bool {
				if len(a) != len(b) {
					return false
				}
				for k, facesA := range a {
					facesB, ok := b[k]
					if !ok || len(facesA) != len(facesB) {
						return false
					}
					for i := range facesA {
						if facesA[i] == nil || facesB[i] == nil {
							if facesA[i] != facesB[i] {
								return false
							}
						} else if facesA[i].Id != facesB[i].Id || facesA[i].ImageId != facesB[i].ImageId {
							return false
						}
					}
				}
				return true
			}

			if !equalFacesByIdAndImageId(got, tt.want) {
				t.Errorf("taskRepo.GetFacesByImageIds() = %v, want %v", got, tt.want)
			}
		})
	}
}
