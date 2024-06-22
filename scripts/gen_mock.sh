#!/bin/bash

set -ex

mockgen -destination=internal/tests/mocks/mock_tddl.go -package=mocks -source=pkg/tddl/tddl.go TDDL
mockgen -destination=internal/tests/mocks/mock_storage.go -package=mocks -source=pkg/storage/storage.go Storage
mockgen -destination=internal/tests/mocks/mock_cache.go -package=mocks -mock_names=Interface=MockCache -source=pkg/cache/interface.go Interface