package juchoi.template.appfolder.redis;

import lombok.RequiredArgsConstructor;
import org.springframework.data.redis.core.RedisTemplate;

import java.time.Duration;

// RedisConfig에서 @Bean으로 등록
@RequiredArgsConstructor
public class RedisModel {

    private final RedisTemplate<String, Object> redisTemplate;

    public void set(String key, Object value, Duration ttl) { redisTemplate.opsForValue().set(key, value, ttl); }
    public Object get(String key)                           { return redisTemplate.opsForValue().get(key); }
    public void delete(String key)                          { redisTemplate.delete(key); }
    public Boolean exists(String key)                       { return redisTemplate.hasKey(key); }
}
