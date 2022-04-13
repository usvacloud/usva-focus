# usvad-sierra

```
docker run -p 80:8080 mattipaksula/usvad-sierra
```

## oneliner to update and leave running

```shell
docker pull mattipaksula/usvad-sierra \
&& docker rm -f usvad-sierra \
&& docker run -d --name usvad-sierra --rm -p 80:8080 mattipaksula/usvad-sierra \
&& docker logs -f usvad-sierra
```
