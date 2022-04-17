# usva-focus

```
docker run -p 80:8080 mattipaksula/usva-focus daemon
```

## oneliner to update and leave running

```shell
docker pull mattipaksula/usva-focus \
&& docker rm -f usva-focus \
&& docker run -d --name usva-focus --rm -p 80:8080 mattipaksula/usva-focus daemon \
&& docker logs -f usva-focus
```
