import redis
import argparse

def query_state(redis_ip, redis_port):

    r = redis.StrictRedis(host==redis_ip, port=redis_port, db=0)
    for key in r.scan_iter("eth:miners:*"):
        print(key, r.hgetall(key))




def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--ip", default="localhost", type=str, help="Redis IP")


    parser.add_argument("--port", default=6379, type=int, help="Redis Port")


    args = parser.parse_args()


    query_state(args.ip, args.port)


if __name__ == "__main__":
    # query syncing state
    main()
