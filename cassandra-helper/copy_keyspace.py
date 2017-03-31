#!/usr/bin/env python2.7
import pika
import json, os
import magic
import time
import ast
from sys import argv
from cassandra.cluster import Cluster
from cassandra.auth import PlainTextAuthProvider
from cassandra import query
from sets import Set

def print_usage():
	print("USAGE: %s KEYSPACE_FROM KEYSPACE_TO TABLE SELECTOR CLUSTER_IPS USERNAME PASSWORD" % argv[0])
	print("e.g.:\n%s holmes_totem holmes results \"service_name = 'yara'\" \"['10.0.4.80','10.0.4.81','10.0.4.82']\" cassandra password" % argv[0])
	exit(-1)

if len(argv) != 8:
	print_usage()

keyspace_from = argv[1]
keyspace_to   = argv[2]
table         = argv[3]
selection     = argv[4]
cluster_ips   = ast.literal_eval(argv[5])
username      = argv[6]
password      = argv[7]

if type(cluster_ips) != list:
	print("ERROR: CLUSTER_IPS must be a list!")
	print_usage()

print("Copying from keyspace '%s' to '%s' on cluster %s: Table '%s' where \"%s\".\n\nContinue? [yn]" % (keyspace_from, keyspace_to, cluster_ips, table, selection))
c = ""
while c != "y":
	c = raw_input()
	if c == 'n':
		print("Aborted")
		exit(-1)

ap = PlainTextAuthProvider(username=username, password=password)
cluster = Cluster(cluster_ips, auth_provider=ap)

sess_get = cluster.connect(keyspace_from)
sess_insert = cluster.connect(keyspace_to)
sess_get.row_factory = query.dict_factory 

rows = sess_get.execute("SELECT * FROM %s WHERE %s;" % (table, selection))
i = 0
for r in rows:
	i += 1
	keys = []
	vals = []
	for k in r:
		keys.append("%s" % str(k))
		vals.append("%%(%s)s" % str(k))
	
	insert_stmt = "INSERT INTO %s (%s) VALUES (%s)" % (table, ",".join(keys), ",".join(vals))
	sess_insert.execute(insert_stmt, r)
	print("Copied %d" % (i))
print("=======")
print("Copied %d entries" % i)
