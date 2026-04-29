package juchoi.template.appfolder;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

// scanBasePackages를 config/로 제한 — 모든 빈은 config/ 내 @Configuration 클래스에서 @Bean으로 등록
@SpringBootApplication(scanBasePackages = "juchoi.template.appfolder.config")
public class AppfolderApplication {

    public static void main(String[] args) {
        SpringApplication.run(AppfolderApplication.class, args);
    }
}
