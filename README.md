# usvad-galant

```
docker run -p 80:8080 mattipaksula/usvad-galant
```

## oneliner to update and leave running

```shell
docker pull mattipaksula/usvad-galant \
&& docker rm -f usvad-galant \
&& docker run -d --name usvad-galant --rm -p 80:8080 mattipaksula/usvad-galant \
&& docker logs -f usvad-galant
```
