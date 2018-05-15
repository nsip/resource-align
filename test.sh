echo '\ncurl -i -X GET http://localhost:1576/align/\n\n'
curl -i -X GET http://localhost:1576/align
echo '\ncurl -i -X GET http://localhost:1576/align?area=Science&year=6,7,8\n\n'
curl -i -X GET "http://localhost:1576/align?area=Science&year=6,7,8"
