FROM alpine
LABEL maintainer = "Webank CTB Team"

ENV APP_HOME=/app/monitor
ENV APP_CONF=$APP_HOME/conf
ENV LOG_PATH=$APP_HOME/logs
ENV PUBLIC_PATH=$APP_HOME/public

RUN apk add ca-certificates
RUN mkdir -p $APP_HOME $APP_CONF $LOG_PATH $PUBLIC_PATH

ADD monitor-server/monitor-server $APP_HOME/
ADD monitor-ui/dist $PUBLIC_PATH/
ADD build/start.sh $APP_HOME/
ADD build/stop.sh $APP_HOME/
ADD monitor-server/conf $APP_CONF/

RUN chmod +x $APP_HOME/*.*

WORKDIR $APP_HOME

ENTRYPOINT ["/bin/sh", "start.sh"]