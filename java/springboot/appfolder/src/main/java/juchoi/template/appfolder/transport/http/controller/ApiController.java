package juchoi.template.appfolder.transport.http.controller;

import juchoi.template.appfolder.eventbus.EventBus;
import juchoi.template.appfolder.eventbus.event.TcpSendEvent;
import juchoi.template.appfolder.service.ApiService;
import juchoi.template.appfolder.transport.http.request.SampleRequest;
import juchoi.template.appfolder.transport.http.response.BaseResponse;
import juchoi.template.appfolder.transport.http.response.SampleResponse;
import juchoi.template.appfolder.transport.http.util.JwtUtil;
import jakarta.servlet.http.Cookie;
import jakarta.servlet.http.HttpServletResponse;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.concurrent.TimeUnit;

// @RestController는 유지 — Spring MVC가 RequestMapping 인식에 사용
// 빈 등록은 WebConfig에서 @Bean으로 처리
@RestController
@RequiredArgsConstructor
public class ApiController {

    private final ApiService apiService;
    private final EventBus eventBus;
    private final JwtUtil jwtUtil;

    @GetMapping("/test")
    public ResponseEntity<String> test(HttpServletResponse response) throws Exception {
        String token = jwtUtil.create(1, "name");
        Cookie cookie = new Cookie("jwt", token);
        cookie.setHttpOnly(true);
        cookie.setPath("/");
        cookie.setMaxAge(60 * 60);
        response.addCookie(cookie);
        return ResponseEntity.ok(apiService.test());
    }

    @GetMapping("/jwt/test")
    public ResponseEntity<String> jwtTest() {
        return ResponseEntity.ok("success jwt test");
    }

    @PostMapping("/sample")
    public ResponseEntity<BaseResponse<SampleResponse>> sample(@RequestBody SampleRequest request) {
        return ResponseEntity.ok(BaseResponse.ok(apiService.handleSample(request)));
    }

    @GetMapping("/send-work")
    public ResponseEntity<String> sendWork() {
        TcpSendEvent event = TcpSendEvent.noReply(1, (byte) 0x11, "hello world, no reply".getBytes());
        eventBus.publish("tcpclient.send", event);
        return ResponseEntity.ok("sent");
    }

    @GetMapping("/send-work-reply")
    public ResponseEntity<String> sendWorkReply() throws Exception {
        TcpSendEvent event = TcpSendEvent.withReply(1, (byte) 0x12, "hello world, with reply".getBytes(), (byte) 0x22);
        eventBus.publish("tcpclient.send", event);
        byte[] reply = event.future().get(10, TimeUnit.SECONDS);
        return ResponseEntity.ok(new String(reply));
    }
}
