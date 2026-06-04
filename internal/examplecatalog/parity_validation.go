package examplecatalog

// ParityFixValidationTarget records the cross-fixture cluster(s) that must be
// checked when accepting a parity fix for a catalog case.
type ParityFixValidationTarget struct {
	CaseID     string
	ClusterIDs []string
}

var parityFixValidationTargets = []ParityFixValidationTarget{
	{CaseID: "fill_basic", ClusterIDs: []string{ValidationClusterPathStrokeMarker}},
	{CaseID: "fill_stacked", ClusterIDs: []string{ValidationClusterPathStrokeMarker}},
	{CaseID: "errorbar_basic", ClusterIDs: []string{ValidationClusterPathStrokeMarker}},
	{CaseID: "boxplot_basic", ClusterIDs: []string{ValidationClusterStatsSpecialty, ValidationClusterPathStrokeMarker}},
	{CaseID: "text_labels_strict", ClusterIDs: []string{ValidationClusterLayoutText, ValidationClusterMathText}},
	{CaseID: "mathtext_basic", ClusterIDs: []string{ValidationClusterMathText}},
	{CaseID: "mathtext_fractions", ClusterIDs: []string{ValidationClusterMathText}},
	{CaseID: "mathtext_integrals", ClusterIDs: []string{ValidationClusterMathText}},
	{CaseID: "mathtext_matrices", ClusterIDs: []string{ValidationClusterMathText}},
	{CaseID: "mathtext_inline_labels", ClusterIDs: []string{ValidationClusterMathText}},
	{CaseID: "image_heatmap", ClusterIDs: []string{ValidationClusterImageMeshColorbar}},
	{CaseID: "imshow_clipped", ClusterIDs: []string{ValidationClusterImageMeshColorbar}},
	{CaseID: "imshow_transformed", ClusterIDs: []string{ValidationClusterImageMeshColorbar}},
	{CaseID: "spy_marker", ClusterIDs: []string{ValidationClusterPathStrokeMarker, ValidationClusterImageMeshColorbar}},
	{CaseID: "spy_image", ClusterIDs: []string{ValidationClusterImageMeshColorbar}},
	{CaseID: "axes_top_right_inverted", ClusterIDs: []string{ValidationClusterLayoutText}},
	{CaseID: "axes_control_surface", ClusterIDs: []string{ValidationClusterLayoutText}},
	{CaseID: "transform_coordinates", ClusterIDs: []string{ValidationClusterLayoutText}},
	{CaseID: "figure_labels_composition", ClusterIDs: []string{ValidationClusterLayoutText}},
	{CaseID: "colorbar_composition", ClusterIDs: []string{ValidationClusterLayoutText, ValidationClusterImageMeshColorbar}},
	{CaseID: "annotation_composition", ClusterIDs: []string{ValidationClusterLayoutText}},
	{CaseID: "legend_layout_matrix", ClusterIDs: []string{ValidationClusterLayoutText}},
	{CaseID: "text_annotation_matrix", ClusterIDs: []string{ValidationClusterLayoutText}},
	{CaseID: "patch_showcase", ClusterIDs: []string{ValidationClusterPathStrokeMarker}},
	{CaseID: "patch_style_matrix", ClusterIDs: []string{ValidationClusterPathStrokeMarker}},
	{CaseID: "mesh_contour_tri", ClusterIDs: []string{ValidationClusterImageMeshColorbar, ValidationClusterContour}},
	{CaseID: "plot_variants", ClusterIDs: []string{ValidationClusterPathStrokeMarker}},
	{CaseID: "spectrum_variants", ClusterIDs: []string{ValidationClusterSignalUnitsVector}},
	{CaseID: "specialty_depth", ClusterIDs: []string{ValidationClusterStatsSpecialty, ValidationClusterPathStrokeMarker}},
	{CaseID: "stem_plot", ClusterIDs: []string{ValidationClusterStatsSpecialty, ValidationClusterPathStrokeMarker}},
	{CaseID: "specialty_artists", ClusterIDs: []string{ValidationClusterStatsSpecialty}},
	{CaseID: "units_overview", ClusterIDs: []string{ValidationClusterSignalUnitsVector}},
	{CaseID: "units_dates", ClusterIDs: []string{ValidationClusterSignalUnitsVector, ValidationClusterPathStrokeMarker}},
	{CaseID: "units_categories", ClusterIDs: []string{ValidationClusterSignalUnitsVector}},
	{CaseID: "units_custom_converter", ClusterIDs: []string{ValidationClusterSignalUnitsVector}},
	{CaseID: "vector_fields", ClusterIDs: []string{ValidationClusterSignalUnitsVector, ValidationClusterPathStrokeMarker}},
	{CaseID: "polar_axes", ClusterIDs: []string{ValidationClusterProjection3D}},
	{CaseID: "geo_mollweide_axes", ClusterIDs: []string{ValidationClusterProjection3D}},
	{CaseID: "geo_aitoff_axes", ClusterIDs: []string{ValidationClusterProjection3D}},
	{CaseID: "geo_hammer_axes", ClusterIDs: []string{ValidationClusterProjection3D}},
	{CaseID: "geo_lambert_axes", ClusterIDs: []string{ValidationClusterProjection3D}},
	{CaseID: "radar_basic", ClusterIDs: []string{ValidationClusterProjection3D}},
	{CaseID: "skewt_basic", ClusterIDs: []string{ValidationClusterProjection3D}},
	{CaseID: "mplot3d_basic", ClusterIDs: []string{ValidationClusterProjection3D}},
	{CaseID: "mplot3d_terrain", ClusterIDs: []string{ValidationClusterProjection3D}},
	{CaseID: "mplot3d_plot3d", ClusterIDs: []string{ValidationClusterProjection3D}},
	{CaseID: "mplot3d_scatter3d", ClusterIDs: []string{ValidationClusterProjection3D}},
	{CaseID: "mplot3d_surface3d", ClusterIDs: []string{ValidationClusterProjection3D}},
	{CaseID: "mplot3d_wire3d", ClusterIDs: []string{ValidationClusterProjection3D}},
	{CaseID: "mplot3d_trisurf3d", ClusterIDs: []string{ValidationClusterProjection3D}},
	{CaseID: "mplot3d_bar3d", ClusterIDs: []string{ValidationClusterProjection3D}},
	{CaseID: "mplot3d_voxels", ClusterIDs: []string{ValidationClusterProjection3D}},
	{CaseID: "mplot3d_quiver3d", ClusterIDs: []string{ValidationClusterProjection3D}},
	{CaseID: "mplot3d_errorbar3d", ClusterIDs: []string{ValidationClusterProjection3D}},
	{CaseID: "mplot3d_stem3d", ClusterIDs: []string{ValidationClusterProjection3D}},
	{CaseID: "mplot3d_fill_between3d", ClusterIDs: []string{ValidationClusterProjection3D}},
	{CaseID: "unstructured_showcase", ClusterIDs: []string{ValidationClusterContour, ValidationClusterImageMeshColorbar}},
	{CaseID: "arrays_showcase", ClusterIDs: []string{ValidationClusterContour, ValidationClusterImageMeshColorbar}},
	{CaseID: "axisartist_showcase", ClusterIDs: []string{ValidationClusterLayoutText}},
	{CaseID: "axes_grid1_showcase", ClusterIDs: []string{ValidationClusterLayoutText}},
	{CaseID: "pcolor_flat", ClusterIDs: []string{ValidationClusterImageMeshColorbar, ValidationClusterContour}},
	{CaseID: "pcolormesh_gouraud", ClusterIDs: []string{ValidationClusterImageMeshColorbar, ValidationClusterContour}},
	{CaseID: "hist2d_weighted_density", ClusterIDs: []string{ValidationClusterImageMeshColorbar}},
	{CaseID: "asinh_norm_image", ClusterIDs: []string{ValidationClusterImageMeshColorbar}},
	{CaseID: "boundarynorm_pcolormesh", ClusterIDs: []string{ValidationClusterImageMeshColorbar}},
	{CaseID: "collection_mutable_scalarmap", ClusterIDs: []string{ValidationClusterImageMeshColorbar}},
	{CaseID: "colorbar_boundary_values", ClusterIDs: []string{ValidationClusterImageMeshColorbar}},
	{CaseID: "colorbar_horizontal_ticks", ClusterIDs: []string{ValidationClusterImageMeshColorbar}},
	{CaseID: "lognorm_imshow", ClusterIDs: []string{ValidationClusterImageMeshColorbar}},
	{CaseID: "twoslope_norm_image", ClusterIDs: []string{ValidationClusterImageMeshColorbar}},
	{CaseID: "colorbar_extensions", ClusterIDs: []string{ValidationClusterImageMeshColorbar}},
}

// ParityFixValidationTargets returns the catalog case-to-cluster validation map.
func ParityFixValidationTargets() []ParityFixValidationTarget {
	out := make([]ParityFixValidationTarget, len(parityFixValidationTargets))
	for i := range parityFixValidationTargets {
		out[i] = parityFixValidationTargets[i]
		out[i].ClusterIDs = append([]string(nil), parityFixValidationTargets[i].ClusterIDs...)
	}
	return out
}
