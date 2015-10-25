go build
mv kubernetes_management docker/kubernetes_management/
find ! -wholename './docker/*' ! -wholename './docker' ! -wholename '.' -exec rm -rf {} +
mv docker/version version
mv docker/environment environment
