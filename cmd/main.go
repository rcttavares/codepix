package main

import (
	"os"

	"github.com/jinzhu/gorm"
	"github.com/rcttavares/codepix/application/grpc"
	"github.com/rcttavares/codepix/infrastructure/db"
)

var database *gorm.DB

func main() {
	database = db.ConnectDB(os.Getenv("env"))
	grpc.StartGrpcServer(database, 50051)
}
