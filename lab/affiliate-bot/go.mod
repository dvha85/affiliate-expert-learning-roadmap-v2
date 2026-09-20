module github.com/dvha85/affiliate-expert-learning-roadmap-v2/lab/affiliate-bot

go 1.27

require (
	github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts v0.0.0
	github.com/dvha85/affiliate-expert-learning-roadmap-v2/core v0.0.0
	golang.org/x/sys v0.46.0
)

require (
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.3 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts => ../../contracts

replace github.com/dvha85/affiliate-expert-learning-roadmap-v2/core => ../../core
