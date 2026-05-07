package juchoi.template.appfolder.transport.http.filter;

import jakarta.servlet.*;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import lombok.extern.slf4j.Slf4j;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Set;

@Slf4j
public class SpaFilter implements Filter {

    private static final Set<String> BLOCKED_EXTENSIONS = Set.of(".git", ".ini", ".svg", ".txt");

    private final String urlPrefix;   // "" → root, "/spatest" → prefix
    private final Path staticDir;
    private final Set<String> redirectPaths;  // root 전용: 해당 경로 → "/" 리다이렉트

    public SpaFilter(String urlPrefix, String staticPath, Set<String> redirectPaths) {
        this.urlPrefix = urlPrefix;
        this.staticDir = Path.of(staticPath).toAbsolutePath().normalize();
        this.redirectPaths = redirectPaths;
    }

    @Override
    public void doFilter(ServletRequest req, ServletResponse res, FilterChain chain)
            throws IOException, ServletException {

        HttpServletRequest request = (HttpServletRequest) req;
        HttpServletResponse response = (HttpServletResponse) res;

        String uri = request.getRequestURI();

        // prefix 모드: 해당 prefix로 시작하지 않으면 통과
        if (!urlPrefix.isEmpty() && !uri.startsWith(urlPrefix)) {
            chain.doFilter(request, response);
            return;
        }

        String relativePath = urlPrefix.isEmpty() ? uri : uri.substring(urlPrefix.length());
        if (relativePath.isEmpty() || relativePath.equals("/")) {
            relativePath = "/index.html";
        }

        // 민감한 확장자 차단
        String ext = extension(relativePath);
        if (BLOCKED_EXTENSIONS.contains(ext)) {
            response.sendError(HttpServletResponse.SC_NOT_FOUND);
            return;
        }

        // 리다이렉트 (root 전용)
        if (urlPrefix.isEmpty() && redirectPaths.contains(relativePath)) {
            response.sendRedirect("/");
            return;
        }

        // path traversal 방지
        Path filePath = staticDir
                .resolve(relativePath.startsWith("/") ? relativePath.substring(1) : relativePath)
                .normalize();
        if (!filePath.startsWith(staticDir)) {
            response.sendError(HttpServletResponse.SC_NOT_FOUND);
            return;
        }

        if (!Files.exists(filePath) || !Files.isRegularFile(filePath)) {
            if (!urlPrefix.isEmpty()) {
                // prefix 모드: 파일 없으면 404
                response.sendError(HttpServletResponse.SC_NOT_FOUND);
                return;
            }
            // root 모드: 파일 없으면 API 라우터로 통과
            chain.doFilter(request, response);
            return;
        }

        String contentType = Files.probeContentType(filePath);
        response.setContentType(contentType != null ? contentType : "application/octet-stream");
        response.setContentLengthLong(Files.size(filePath));
        Files.copy(filePath, response.getOutputStream());
    }

    private String extension(String path) {
        int dot = path.lastIndexOf('.');
        int slash = path.lastIndexOf('/');
        return dot > slash ? path.substring(dot) : "";
    }
}
