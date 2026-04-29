package juchoi.template.appfolder.config;

import juchoi.template.appfolder.eventbus.EventBus;
import juchoi.template.appfolder.service.ApiService;
import juchoi.template.appfolder.transport.http.controller.ApiController;
import juchoi.template.appfolder.transport.http.controller.SseController;
import juchoi.template.appfolder.transport.http.filter.JwtFilter;
import juchoi.template.appfolder.transport.http.filter.LogFilter;
import juchoi.template.appfolder.transport.http.util.JwtUtil;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.web.servlet.FilterRegistrationBean;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.cors.CorsConfiguration;
import org.springframework.web.cors.UrlBasedCorsConfigurationSource;
import org.springframework.web.filter.CorsFilter;

import java.util.List;

// HTTP 전송 계층 빈 등록 — 컨트롤러, 필터, JWT 유틸
@Configuration
public class WebConfig {

    @Bean
    public JwtUtil jwtUtil(@Value("${jwt.secret}") String secret) {
        return new JwtUtil(secret);
    }

    @Bean
    public LogFilter logFilter() {
        return new LogFilter();
    }

    @Bean
    public JwtFilter jwtFilter(@Value("${jwt.secret}") String secret) {
        return new JwtFilter(secret);
    }

    // /jwt/* 경로에만 JWT 필터 적용
    @Bean
    public FilterRegistrationBean<JwtFilter> jwtFilterRegistration(JwtFilter jwtFilter) {
        FilterRegistrationBean<JwtFilter> bean = new FilterRegistrationBean<>(jwtFilter);
        bean.addUrlPatterns("/jwt/*");
        bean.setOrder(1);
        return bean;
    }

    @Bean
    public FilterRegistrationBean<CorsFilter> corsFilter() {
        CorsConfiguration config = new CorsConfiguration();
        config.setAllowedOriginPatterns(List.of("*"));
        config.setAllowedMethods(List.of("GET", "POST", "PUT", "DELETE", "OPTIONS"));
        config.setAllowedHeaders(List.of("*"));
        config.setAllowCredentials(true);

        UrlBasedCorsConfigurationSource source = new UrlBasedCorsConfigurationSource();
        source.registerCorsConfiguration("/**", config);

        FilterRegistrationBean<CorsFilter> bean = new FilterRegistrationBean<>(new CorsFilter(source));
        bean.setOrder(0);
        return bean;
    }

    // @RestController가 클래스에 남아있어야 Spring MVC가 RequestMapping을 인식함
    @Bean
    public ApiController apiController(ApiService apiService, EventBus eventBus, JwtUtil jwtUtil) {
        return new ApiController(apiService, eventBus, jwtUtil);
    }

    @Bean
    public SseController sseController(EventBus eventBus) {
        return new SseController(eventBus);
    }
}
