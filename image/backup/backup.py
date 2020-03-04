#coding: utf8
import urllib2, json, redis, time, os

dashboard = os.getenv("DASHBOARD_ADDR")
url = "http://%s/topom/stats/" % (dashboard)
try:
    res = urllib2.urlopen(url)
    if res.status_code  == 200:
       text = json.loads(res.text)
       models = text['group']['models']
       group_nums = len(models)
       for i in range(group_nums):
            print "read group ",i
            redis_ip = models[i]['servers'][1]['server'].split(':',1)[0]
            print "backup begining.... connect" , redis_ip
            try:
                r = redis.Redis(host=redis_ip, port=6379, db=0)
                print redis_ip
                r.bgsave()
                print "backup finished !"
            except Exception, e:
                print "backup failed:", e
            time.sleep(5)
    else:
        print "get topo failed"
except Exception, e:
    print "request failed:" ,e
