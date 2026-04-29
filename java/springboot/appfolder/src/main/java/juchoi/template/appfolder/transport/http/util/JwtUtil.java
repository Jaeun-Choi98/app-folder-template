package juchoi.template.appfolder.transport.http.util;

import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.security.Keys;
import lombok.RequiredArgsConstructor;

import java.nio.charset.StandardCharsets;
import java.util.Date;

// WebConfig에서 @Bean으로 등록 — @Value는 config 클래스가 주입
@RequiredArgsConstructor
public class JwtUtil {

    private final String secret;

    public String create(long userId, String name) {
        return Jwts.builder()
                .subject(String.valueOf(userId))
                .claim("name", name)
                .issuedAt(new Date())
                .expiration(new Date(System.currentTimeMillis() + 3_600_000L))
                .signWith(Keys.hmacShaKeyFor(secret.getBytes(StandardCharsets.UTF_8)))
                .compact();
    }
}
