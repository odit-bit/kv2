# KV2
key value store application that implement redis protocol

## Quickstart 

```
git clone https://github.com/odit-bit/kv2
```  
```
cd kv2
```
### server

#### start
it is main entry for server 
```
go run ./cli/server run --debug
```  


#### or docker 
```
docker run --name kv2 -d oditbit/kv2:1.0
```
#### or compose
``` 
docker compose up -d
```

server default bind port `6969`

### client

#### cli  
- set 
```
go run ./cli/client SET key value
```  
- get  
```
go run ./cli/client GET key
```

***

it can use redis client in go project, see `/example/redis`

```	golang
import "github.com/odit-bit/kv2/client"

func main() {
	cli := redis.NewClient(&redis.Options{
		Addr:             "localhost:6969",
		DisableIndentity: true,
	})
	defer cli.Close()

    //set command
    res := cli.Set(context.Background(), "key", "value", 0)
    if res.Err() != nil {
        log.Println(res.Err())
        return
    }

    //get command
    res := cli.Get(context.Background(), test.key)
    if res.Err() != nil {
        log.Println(res.Err())
        return 
    }

    fmt.Println(res.String())

}
```
for more example see `/example` directory
***
### Command
#### SET
    SET key value [EX second | PX milisecond]

`EX` make key expired in second while `PX` in milisecond

#### GET
    GET key



***
### Configuration
Default maximum capacity (ram) is 512 mb, it can change with flag `--size byte`, if max is reached and  overflow happen it will drop some to make room for new entry due to the behaviour of underlying cache implementation.
***

<!-- ## License

this source code is available under an MIT [License](https://github.com/odit-bit/kv2/blob/master/LICENSE.md). -->