## Python ##

```
python3 -m venv ~/.virtualenvs/k720
source ~/.virtualenvs/k720/bin/activate
pip3 install pyserial
mkdir k720_pylib
cd k720_pylib/
pip freeze > requirements.txt
cd ..
```

## Golang ##

```
go install -v go.bug.st/serial@latest
```
```
mkdir k720_golib
cd k720_golib/
go mod init k720
go get go.bug.st/serial@latest
cd ..
```
```
go mod init main_k720
go work init . ./k
```
