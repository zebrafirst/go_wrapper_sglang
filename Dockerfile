FROM artifacts.iflytek.com/docker-private/aipaas/aiges-build:2.9.13-th3api as builder

COPY . /home/AIGES/go_wrapper_sglang
RUN bash /home/AIGES/build.wrapper.sh

FROM artifacts.iflytek.com/docker-private/atp/vllm_oai_wrapper:v1.0.27

WORKDIR /home/aiges
COPY --from=builder /home/AIGES/bin/libwrapper.so /home/aiges
RUN chmod 755 /home/aiges/libwrapper.so

COPY ./aiservice_2.9.11.5.bin/AIservice /home/aiges
COPY ./aiservice_2.9.11.5.bin/lib /home/aiges/library
