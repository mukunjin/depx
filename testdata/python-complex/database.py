from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker, declarative_base
import redis

# 数据库配置
DATABASE_URL = "postgresql://user:pass@localhost/db"
engine = create_engine(DATABASE_URL)
SessionLocal = sessionmaker(bind=engine)
Base = declarative_base()

# Redis 配置
redis_client = redis.Redis(host='localhost', port=6379, decode_responses=True)

def get_db_session():
    return SessionLocal()

def cache_key(prefix, id):
    return f"{prefix}:{id}"

def get_from_cache(key):
    return redis_client.get(key)

def set_to_cache(key, value, ttl=3600):
    redis_client.setex(key, ttl, value)
