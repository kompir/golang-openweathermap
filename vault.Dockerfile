FROM vault:latest

#COPY ./run.sh /run.sh
#RUN chmod +x /run.sh
#ENTRYPOINT [ "sh", "/run.sh"]
#RUN sh /run.sh

#add sercrets from cmd
RUN export VAULT_ADDR=http://0.0.0.0:8200
RUN export VAULT_TOKEN=myroot
RUN vault kv put secret/database username=admin password=12345678