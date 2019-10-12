gcc -c device_query.c -o device_query.o -lcndev -I ${CNDEV_HOME}/include -L ${CNDEV_HOME}/lib -std=gnu99 

ar -crs libdeviceq.a device_query.o
