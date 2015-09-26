FROM scratch
MAINTAINER Nick Owens <mischief@offblast.org> (@mischief)

ENV PATH=/bin
ENV KITE_HOME=/kite
VOLUME /kite
EXPOSE 4000

ADD bin/* /bin/

CMD ["glenda"]

