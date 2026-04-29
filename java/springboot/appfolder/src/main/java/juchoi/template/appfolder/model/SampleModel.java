package juchoi.template.appfolder.model;

import lombok.Getter;
import lombok.RequiredArgsConstructor;

// Mirrors Go's internal/model/sample.go — domain model passed between layers.
@Getter
@RequiredArgsConstructor
public class SampleModel {
    private final String name;
    private final int value;
}
