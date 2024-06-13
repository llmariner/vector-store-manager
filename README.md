# vector-store-manager
Vector Store Manager

## Running integration test with Milvus server
- `kubectl port-forward -n milvus services/milvus 19530:19530`
- Run `make test-integration`.
