package juchoi.template.appfolder.service;

import juchoi.template.appfolder.transport.http.request.SampleRequest;
import juchoi.template.appfolder.transport.http.response.SampleResponse;

// Mirrors Go's service.APIServcieInterface
public interface ApiService {
    String test();
    SampleResponse handleSample(SampleRequest request);
}
