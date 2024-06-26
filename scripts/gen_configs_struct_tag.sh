#!/bin/bash

set -ex

cmd="gomodifytags -all -add-tags json,yaml,mapstructure -w"

${cmd} -file configs/server_config.go
${cmd} -file configs/log_config.go
${cmd} -file configs/tddl_config.go
${cmd} -file configs/cache_config.go
${cmd} -file configs/mysql_config.go