FROM artifacts.iflytek.com/docker-private/aipaas/aiges-build:2.9.14.3-otlp as builder

COPY . /home/AIGES/go_wrapper_sglang
COPY ./build-so.sh /home/AIGES/build-so.sh
RUN bash /home/AIGES/build-so.sh

FROM artifacts.iflytek.com/docker-private/atp/vllm_oai_wrapper:v1.0.27

WORKDIR /home/aiges
COPY --from=builder /home/AIGES/bin/libwrapper.so /home/aiges
RUN chmod 755 /home/aiges/libwrapper.so

COPY ./build/aiservice_2.9.14.3-otlp.bin/AIservice /home/aiges
COPY ./build/aiservice_2.9.14.3-otlp.bin/lib /home/aiges/library
RUN chmod 755 /home/aiges/AIservice
