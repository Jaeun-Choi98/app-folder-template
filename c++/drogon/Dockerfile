FROM ubuntu:22.04 AS builder

RUN apt-get update && apt-get install -y \
    build-essential cmake git \
    libjsoncpp-dev uuid-dev zlib1g-dev \
    libssl-dev libmariadb-dev \
    libhiredis-dev \
    && rm -rf /var/lib/apt/lists/*

# Drogon
RUN git clone --depth 1 https://github.com/drogonframework/drogon.git /opt/drogon && \
    cd /opt/drogon && git submodule update --init && \
    mkdir build && cd build && \
    cmake .. -DCMAKE_BUILD_TYPE=Release && \
    make -j$(nproc) && make install

# SOCI
RUN git clone --depth 1 https://github.com/SOCI/soci.git /opt/soci && \
    cd /opt/soci && mkdir build && cd build && \
    cmake .. -DCMAKE_BUILD_TYPE=Release -DWITH_MYSQL=ON && \
    make -j$(nproc) && make install

# standalone asio
RUN git clone --depth 1 https://github.com/chriskohlhoff/asio.git /opt/asio
ENV CPLUS_INCLUDE_PATH=/opt/asio/asio/include

WORKDIR /app
COPY . .
RUN mkdir build && cd build && \
    cmake .. -DCMAKE_BUILD_TYPE=Release && \
    make -j$(nproc)

FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    libjsoncpp25 libssl3 libmariadb3 libhiredis0.14 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /usr/local/lib /usr/local/lib
COPY --from=builder /app/build/appfolder /usr/local/bin/appfolder

RUN ldconfig
WORKDIR /app
COPY cmd/server/envfile_here.txt .

EXPOSE 8080 9090
CMD ["appfolder"]
