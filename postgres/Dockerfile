FROM pgvector/pgvector:pg18-trixie

# Install build dependencies, compile extensions, and clean up in a single layer to compress image size
RUN apt-get update && apt-get install -y --no-install-recommends \
    postgresql-server-dev-18 \
    make \
    gcc \
    curl \
    wget \
    ca-certificates \
    bzip2 \
    # Install scws (Dependency for zhparser)
    && wget -q -O - http://www.xunsearch.com/scws/down/scws-1.2.3.tar.bz2 | tar xjf - \
    && cd scws-1.2.3 \
    && ./configure \
    && make install \
    && cd .. \
    && rm -rf scws-1.2.3 \
    # Install zhparser
    && wget -q -O - https://github.com/amutu/zhparser/archive/master.tar.gz | tar xzf - \
    && cd zhparser-master \
    && SCWS_HOME=/usr/local make \
    && make install \
    && cd .. \
    && rm -rf zhparser-master \
    # Install pgmq
    && curl -LO https://github.com/tembo-io/pgmq/archive/refs/tags/v1.4.5.tar.gz \
    && tar xzf v1.4.5.tar.gz \
    && cd pgmq-1.4.5/pgmq-extension \
    && make \
    && make install \
    && cd ../.. \
    && rm -rf v1.4.5.tar.gz pgmq-1.4.5 \
    # Strip libraries to save space
    && strip /usr/lib/postgresql/18/lib/zhparser.so \
    && strip /usr/lib/postgresql/18/lib/pgmq.so \
    && strip /usr/local/lib/libscws.so \
    # Clean up build dependencies
    && apt-get remove -y \
    postgresql-server-dev-18 \
    make \
    gcc \
    bzip2 \
    && apt-get autoremove -y \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Auto enable pg_stat_statements by adding it to shared_preload_libraries
RUN echo "shared_preload_libraries = 'pg_stat_statements'" >> /usr/share/postgresql/postgresql.conf.sample

# Copy initialization script to auto-enable extensions on startup
COPY init-extensions.sql /docker-entrypoint-initdb.d/