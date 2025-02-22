package repository

import (
	"context"
	"time"

	"github.com/zhashkevych/courses-backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type StudentsRepo struct {
	db *mongo.Collection
}

func NewStudentsRepo(db *mongo.Database) *StudentsRepo {
	return &StudentsRepo{
		db: db.Collection(studentsCollection),
	}
}

func (r *StudentsRepo) Create(ctx context.Context, student domain.Student) error {
	_, err := r.db.InsertOne(ctx, student)
	return err
}

func (r *StudentsRepo) GetByCredentials(ctx context.Context, schoolId primitive.ObjectID, email, password string) (domain.Student, error) {
	var student domain.Student
	if err := r.db.FindOne(ctx, bson.M{"email": email, "password": password, "schoolId": schoolId, "verification.verified": true}).
		Decode(&student); err != nil {
		if err == mongo.ErrNoDocuments {
			return domain.Student{}, ErrUserNotFound
		}

		return domain.Student{}, err
	}

	return student, nil
}

func (r *StudentsRepo) GetByRefreshToken(ctx context.Context, schoolId primitive.ObjectID, refreshToken string) (domain.Student, error) {
	var student domain.Student
	if err := r.db.FindOne(ctx, bson.M{"session.refreshToken": refreshToken, "schoolId": schoolId,
		"session.expiresAt": bson.M{"$gt": time.Now()}}).Decode(&student); err != nil {
		if err == mongo.ErrNoDocuments {
			return domain.Student{}, ErrUserNotFound
		}

		return domain.Student{}, err
	}

	return student, nil
}

func (r *StudentsRepo) GetById(ctx context.Context, id primitive.ObjectID) (domain.Student, error) {
	var student domain.Student
	err := r.db.FindOne(ctx, bson.M{"_id": id, "verification.verified": true}).Decode(&student)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return domain.Student{}, ErrUserNotFound
		}

		return domain.Student{}, err
	}

	return student, nil
}

func (r *StudentsRepo) GetBySchool(ctx context.Context, schoolId primitive.ObjectID) ([]domain.Student, error) {
	cur, err := r.db.Find(ctx, bson.M{"schoolId": schoolId})
	if err != nil {
		return nil, err
	}

	var students []domain.Student
	err = cur.All(ctx, &students)
	return students, err
}

func (r *StudentsRepo) SetSession(ctx context.Context, studentId primitive.ObjectID, session domain.Session) error {
	_, err := r.db.UpdateOne(ctx, bson.M{"_id": studentId}, bson.M{"$set": bson.M{"session": session, "lastVisitAt": time.Now()}})
	return err
}

func (r *StudentsRepo) GiveAccessToCourseAndModule(ctx context.Context, studentId, courseId, moduleId primitive.ObjectID) error {
	_, err := r.db.UpdateOne(ctx, bson.M{"_id": studentId}, bson.M{"$addToSet": bson.M{"availableModules": moduleId, "availableCourses": courseId}})
	return err
}

func (r *StudentsRepo) GiveAccessToCoursesAndModules(ctx context.Context, studentId primitive.ObjectID, courseIds, moduleIds []primitive.ObjectID) error {
	_, err := r.db.UpdateOne(ctx, bson.M{"_id": studentId}, bson.M{"$addToSet": bson.M{"availableModules": bson.M{"$each": moduleIds},
		"availableCourses": bson.M{"$each": courseIds}}})
	return err
}

func (r *StudentsRepo) Verify(ctx context.Context, code string) error {
	res, err := r.db.UpdateOne(ctx,
		bson.M{"verification.code": code},
		bson.M{"$set": bson.M{"verification.verified": true, "verification.code": ""}})
	if err != nil {
		return err
	}

	if res.ModifiedCount == 0 {
		return ErrVerificationCodeInvalid
	}

	return nil
}
