FROM debian:stable-slim

RUN mkdir -p /logs

# COPY source destination
COPY chirpy /bin/chirpy

CMD ["/bin/chirpy"]