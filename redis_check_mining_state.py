import redis
import argparse
import time

def query_state(redis_ip, redis_port, interval):

    r = redis.StrictRedis(host=redis_ip, port=redis_port,password='tothem00n', db=0)
    mydict = {}
    mydict = r.hgetall("minerBoards")
    for k, v in mydict.items():
        print(k, v)


def getPoolBlockCount():
    mydict = {}
    try:
        r = redis.StrictRedis(host="34.220.137.126", port=6379,password='tothem00n', db=0)
        mydict = r.hgetall("minerBoards")
    except Exception:
        print("Can not get the block")
        return mydict
    data = {}
    for k, v in mydict.items():
        data[k.decode("utf-8")] = v.decode("utf-8")
    return data


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--ip", default="34.220.137.126", type=str, help="Redis IP")


    parser.add_argument("--port", default=6379, type=int, help="Redis Port")

    parser.add_argument(
        "-i", "--interval", default=10, type=int, help="Query interval in second"
    )


    args = parser.parse_args()


    #query_state(args.ip, args.port, args.interval
    mydict = getPoolBlockCount()
    for k, v in mydict.items():
        print(k, v)


if __name__ == "__main__":
    # query syncing state
    main()
