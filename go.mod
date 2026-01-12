module github.com/go-anyway/framework-mongodb

go 1.25.4

require (
	github.com/go-anyway/framework-config v1.0.0
	github.com/go-anyway/framework-log v1.0.0
	github.com/go-anyway/framework-trace v1.0.0
	go.mongodb.org/mongo-driver v1.17.6
)

replace (
	github.com/go-anyway/framework-config => ../core/config
	github.com/go-anyway/framework-log => ../core/log
	github.com/go-anyway/framework-trace => ../trace
)
