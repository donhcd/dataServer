from subprocess import call
import json

obj = {"readings":[]}

for i in xrange(100000):
    # if i % 10000 == 0:
    #     print i
    # elif i % 1000 == 0:
    #     print i
    # elif i % 100 == 0:
    #     print i
    t = 1462407200+ i * 100
    obj["readings"].append({"timestamp": t,
                            "blob": "randomblobofdata"+str(i)})
print json.dumps(obj)
# call(["curl", "-X", "POST", "localhost:8080/devices/foobar/insert",
#         "--data",
#         json.dumps(obj)])
