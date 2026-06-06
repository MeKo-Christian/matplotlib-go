package parity

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/cwbudde/matplotlib-go/core"
	showcase_animation_gallery "github.com/cwbudde/matplotlib-go/examples/animation_gallery"
	showcase_annotation_composition "github.com/cwbudde/matplotlib-go/examples/annotation_composition"
	showcase_annotation_legend_offsetbox_gallery "github.com/cwbudde/matplotlib-go/examples/annotation_legend_offsetbox_gallery"
	showcase_arrays_showcase "github.com/cwbudde/matplotlib-go/examples/arrays_showcase"
	showcase_axes_control_surface "github.com/cwbudde/matplotlib-go/examples/axes_control_surface"
	showcase_axes_grid1_showcase "github.com/cwbudde/matplotlib-go/examples/axes_grid1_showcase"
	showcase_axisartist_showcase "github.com/cwbudde/matplotlib-go/examples/axisartist_showcase"
	showcase_bar_basic "github.com/cwbudde/matplotlib-go/examples/bar_basic"
	showcase_basic_line "github.com/cwbudde/matplotlib-go/examples/basic_line"
	showcase_basic_line_labels "github.com/cwbudde/matplotlib-go/examples/basic_line_labels"
	showcase_boxplot_basic "github.com/cwbudde/matplotlib-go/examples/boxplot_basic"
	showcase_colorbar_composition "github.com/cwbudde/matplotlib-go/examples/colorbar_composition"
	showcase_colorbar_variants_gallery "github.com/cwbudde/matplotlib-go/examples/colorbar_variants_gallery"
	showcase_colormap_families_gallery "github.com/cwbudde/matplotlib-go/examples/colormap_families_gallery"
	showcase_errorbar_basic "github.com/cwbudde/matplotlib-go/examples/errorbar_basic"
	showcase_geo_mollweide_axes "github.com/cwbudde/matplotlib-go/examples/geo_mollweide_axes"
	showcase_gridspec_composition "github.com/cwbudde/matplotlib-go/examples/gridspec_composition"
	showcase_hist_basic "github.com/cwbudde/matplotlib-go/examples/hist_basic"
	showcase_image_heatmap "github.com/cwbudde/matplotlib-go/examples/image_heatmap"
	showcase_image_variants_gallery "github.com/cwbudde/matplotlib-go/examples/image_variants_gallery"
	showcase_lines_markers_gallery "github.com/cwbudde/matplotlib-go/examples/lines_markers_gallery"
	showcase_mathtext_gallery "github.com/cwbudde/matplotlib-go/examples/mathtext_gallery"
	showcase_mesh_contour_tri "github.com/cwbudde/matplotlib-go/examples/mesh_contour_tri"
	showcase_mixed_raster_vector "github.com/cwbudde/matplotlib-go/examples/mixed_raster_vector"
	showcase_mplot3d_gallery "github.com/cwbudde/matplotlib-go/examples/mplot3d_gallery"
	showcase_multi_series_basic "github.com/cwbudde/matplotlib-go/examples/multi_series_basic"
	showcase_named_colors_gallery "github.com/cwbudde/matplotlib-go/examples/named_colors_gallery"
	showcase_plot_variants "github.com/cwbudde/matplotlib-go/examples/plot_variants"
	showcase_polar_axes "github.com/cwbudde/matplotlib-go/examples/polar_axes"
	showcase_radar_basic "github.com/cwbudde/matplotlib-go/examples/radar_basic"
	showcase_scatter_basic "github.com/cwbudde/matplotlib-go/examples/scatter_basic"
	showcase_scatter_gallery "github.com/cwbudde/matplotlib-go/examples/scatter_gallery"
	showcase_skewt_basic "github.com/cwbudde/matplotlib-go/examples/skewt_basic"
	showcase_specialty_artists "github.com/cwbudde/matplotlib-go/examples/specialty_artists"
	showcase_stat_variants "github.com/cwbudde/matplotlib-go/examples/stat_variants"
	showcase_text_layout_gallery "github.com/cwbudde/matplotlib-go/examples/text_layout_gallery"
	showcase_units_overview "github.com/cwbudde/matplotlib-go/examples/units_overview"
	showcase_unstructured_showcase "github.com/cwbudde/matplotlib-go/examples/unstructured_showcase"
	showcase_vector_fields "github.com/cwbudde/matplotlib-go/examples/vector_fields"
	"github.com/cwbudde/matplotlib-go/internal/examplecatalog"
	example_annotation_composition "github.com/cwbudde/matplotlib-go/test/parity/annotation_composition"
	example_annotation_legend_offsetbox_gallery "github.com/cwbudde/matplotlib-go/test/parity/annotation_legend_offsetbox_gallery"
	example_arrays_showcase "github.com/cwbudde/matplotlib-go/test/parity/arrays_showcase"
	example_artist_metadata "github.com/cwbudde/matplotlib-go/test/parity/artist_metadata"
	example_asinh_norm_image "github.com/cwbudde/matplotlib-go/test/parity/asinh_norm_image"
	example_axes_control_surface "github.com/cwbudde/matplotlib-go/test/parity/axes_control_surface"
	example_axes_convenience_helpers "github.com/cwbudde/matplotlib-go/test/parity/axes_convenience_helpers"
	example_axes_grid1_showcase "github.com/cwbudde/matplotlib-go/test/parity/axes_grid1_showcase"
	example_axes_option_breadth_17_75_3 "github.com/cwbudde/matplotlib-go/test/parity/axes_option_breadth_17_75_3"
	example_axes_top_right_inverted "github.com/cwbudde/matplotlib-go/test/parity/axes_top_right_inverted"
	example_axisartist_showcase "github.com/cwbudde/matplotlib-go/test/parity/axisartist_showcase"
	example_bar_basic "github.com/cwbudde/matplotlib-go/test/parity/bar_basic"
	example_bar_basic_frame "github.com/cwbudde/matplotlib-go/test/parity/bar_basic_frame"
	example_bar_basic_tick_labels "github.com/cwbudde/matplotlib-go/test/parity/bar_basic_tick_labels"
	example_bar_basic_ticks "github.com/cwbudde/matplotlib-go/test/parity/bar_basic_ticks"
	example_bar_basic_title "github.com/cwbudde/matplotlib-go/test/parity/bar_basic_title"
	example_bar_grouped "github.com/cwbudde/matplotlib-go/test/parity/bar_grouped"
	example_bar_horizontal "github.com/cwbudde/matplotlib-go/test/parity/bar_horizontal"
	example_basic_line "github.com/cwbudde/matplotlib-go/test/parity/basic_line"
	example_basic_line_labels "github.com/cwbudde/matplotlib-go/test/parity/basic_line_labels"
	example_boundarynorm_pcolormesh "github.com/cwbudde/matplotlib-go/test/parity/boundarynorm_pcolormesh"
	example_boxplot_basic "github.com/cwbudde/matplotlib-go/test/parity/boxplot_basic"
	example_clip_path_batch "github.com/cwbudde/matplotlib-go/test/parity/clip_path_batch"
	example_collection_mutable_scalarmap "github.com/cwbudde/matplotlib-go/test/parity/collection_mutable_scalarmap"
	example_colorbar_boundary_values "github.com/cwbudde/matplotlib-go/test/parity/colorbar_boundary_values"
	example_colorbar_composition "github.com/cwbudde/matplotlib-go/test/parity/colorbar_composition"
	example_colorbar_extensions "github.com/cwbudde/matplotlib-go/test/parity/colorbar_extensions"
	example_colorbar_horizontal_ticks "github.com/cwbudde/matplotlib-go/test/parity/colorbar_horizontal_ticks"
	example_colorbar_variants_gallery "github.com/cwbudde/matplotlib-go/test/parity/colorbar_variants_gallery"
	example_colormap_cyclic "github.com/cwbudde/matplotlib-go/test/parity/colormap_cyclic"
	example_colormap_diverging "github.com/cwbudde/matplotlib-go/test/parity/colormap_diverging"
	example_colormap_families_gallery "github.com/cwbudde/matplotlib-go/test/parity/colormap_families_gallery"
	example_colormap_qualitative "github.com/cwbudde/matplotlib-go/test/parity/colormap_qualitative"
	example_dashes "github.com/cwbudde/matplotlib-go/test/parity/dashes"
	example_date_concise_intraday_labels "github.com/cwbudde/matplotlib-go/test/parity/date_concise_intraday_labels"
	example_date_month_year_labels "github.com/cwbudde/matplotlib-go/test/parity/date_month_year_labels"
	example_errorbar_basic "github.com/cwbudde/matplotlib-go/test/parity/errorbar_basic"
	example_figure_labels_composition "github.com/cwbudde/matplotlib-go/test/parity/figure_labels_composition"
	example_fill_basic "github.com/cwbudde/matplotlib-go/test/parity/fill_basic"
	example_fill_between "github.com/cwbudde/matplotlib-go/test/parity/fill_between"
	example_fill_stacked "github.com/cwbudde/matplotlib-go/test/parity/fill_stacked"
	example_formatter_engineering_labels "github.com/cwbudde/matplotlib-go/test/parity/formatter_engineering_labels"
	example_formatter_fixed_null_labels "github.com/cwbudde/matplotlib-go/test/parity/formatter_fixed_null_labels"
	example_formatter_log_mathtext_labels "github.com/cwbudde/matplotlib-go/test/parity/formatter_log_mathtext_labels"
	example_formatter_percent_labels "github.com/cwbudde/matplotlib-go/test/parity/formatter_percent_labels"
	example_formatter_scalar_scientific_labels "github.com/cwbudde/matplotlib-go/test/parity/formatter_scalar_scientific_labels"
	example_geo_aitoff_axes "github.com/cwbudde/matplotlib-go/test/parity/geo_aitoff_axes"
	example_geo_hammer_axes "github.com/cwbudde/matplotlib-go/test/parity/geo_hammer_axes"
	example_geo_lambert_axes "github.com/cwbudde/matplotlib-go/test/parity/geo_lambert_axes"
	example_geo_mollweide_axes "github.com/cwbudde/matplotlib-go/test/parity/geo_mollweide_axes"
	example_gouraud_triangles "github.com/cwbudde/matplotlib-go/test/parity/gouraud_triangles"
	example_gridspec_composition "github.com/cwbudde/matplotlib-go/test/parity/gridspec_composition"
	example_hist2d_weighted_density "github.com/cwbudde/matplotlib-go/test/parity/hist2d_weighted_density"
	example_hist_basic "github.com/cwbudde/matplotlib-go/test/parity/hist_basic"
	example_hist_density "github.com/cwbudde/matplotlib-go/test/parity/hist_density"
	example_hist_strategies "github.com/cwbudde/matplotlib-go/test/parity/hist_strategies"
	example_image_alpha "github.com/cwbudde/matplotlib-go/test/parity/image_alpha"
	example_image_heatmap "github.com/cwbudde/matplotlib-go/test/parity/image_heatmap"
	example_image_variants_gallery "github.com/cwbudde/matplotlib-go/test/parity/image_variants_gallery"
	example_imshow_bicubic "github.com/cwbudde/matplotlib-go/test/parity/imshow_bicubic"
	example_imshow_bilinear "github.com/cwbudde/matplotlib-go/test/parity/imshow_bilinear"
	example_imshow_clipped "github.com/cwbudde/matplotlib-go/test/parity/imshow_clipped"
	example_imshow_interpolation_matrix "github.com/cwbudde/matplotlib-go/test/parity/imshow_interpolation_matrix"
	example_imshow_transformed "github.com/cwbudde/matplotlib-go/test/parity/imshow_transformed"
	example_joins_caps "github.com/cwbudde/matplotlib-go/test/parity/joins_caps"
	example_large_scatter "github.com/cwbudde/matplotlib-go/test/parity/large_scatter"
	example_layout_bbox_helpers "github.com/cwbudde/matplotlib-go/test/parity/layout_bbox_helpers"
	example_legend_layout_matrix "github.com/cwbudde/matplotlib-go/test/parity/legend_layout_matrix"
	example_line2d_markers "github.com/cwbudde/matplotlib-go/test/parity/line2d_markers"
	example_line2d_semantics "github.com/cwbudde/matplotlib-go/test/parity/line2d_semantics"
	example_lines_markers_gallery "github.com/cwbudde/matplotlib-go/test/parity/lines_markers_gallery"
	example_locator_fixed_index_labels "github.com/cwbudde/matplotlib-go/test/parity/locator_fixed_index_labels"
	example_locator_linear_labels "github.com/cwbudde/matplotlib-go/test/parity/locator_linear_labels"
	example_locator_log_minor_threshold_labels "github.com/cwbudde/matplotlib-go/test/parity/locator_log_minor_threshold_labels"
	example_locator_maxn_edge_labels "github.com/cwbudde/matplotlib-go/test/parity/locator_maxn_edge_labels"
	example_lognorm_imshow "github.com/cwbudde/matplotlib-go/test/parity/lognorm_imshow"
	example_mathtext_basic "github.com/cwbudde/matplotlib-go/test/parity/mathtext_basic"
	example_mathtext_fractions "github.com/cwbudde/matplotlib-go/test/parity/mathtext_fractions"
	example_mathtext_gallery "github.com/cwbudde/matplotlib-go/test/parity/mathtext_gallery"
	example_mathtext_inline_labels "github.com/cwbudde/matplotlib-go/test/parity/mathtext_inline_labels"
	example_mathtext_integrals "github.com/cwbudde/matplotlib-go/test/parity/mathtext_integrals"
	example_mathtext_matrices "github.com/cwbudde/matplotlib-go/test/parity/mathtext_matrices"
	example_matshow_basic "github.com/cwbudde/matplotlib-go/test/parity/matshow_basic"
	example_mesh_contour_tri "github.com/cwbudde/matplotlib-go/test/parity/mesh_contour_tri"
	example_mixed_collection "github.com/cwbudde/matplotlib-go/test/parity/mixed_collection"
	example_mixed_raster_vector "github.com/cwbudde/matplotlib-go/test/parity/mixed_raster_vector"
	example_mplot3d_bar2d_zdir "github.com/cwbudde/matplotlib-go/test/parity/mplot3d_bar2d_zdir"
	example_mplot3d_bar3d "github.com/cwbudde/matplotlib-go/test/parity/mplot3d_bar3d"
	example_mplot3d_basic "github.com/cwbudde/matplotlib-go/test/parity/mplot3d_basic"
	example_mplot3d_contour3d "github.com/cwbudde/matplotlib-go/test/parity/mplot3d_contour3d"
	example_mplot3d_contourf3d "github.com/cwbudde/matplotlib-go/test/parity/mplot3d_contourf3d"
	example_mplot3d_errorbar3d "github.com/cwbudde/matplotlib-go/test/parity/mplot3d_errorbar3d"
	example_mplot3d_fill_between3d "github.com/cwbudde/matplotlib-go/test/parity/mplot3d_fill_between3d"
	example_mplot3d_gallery "github.com/cwbudde/matplotlib-go/test/parity/mplot3d_gallery"
	example_mplot3d_plot3d "github.com/cwbudde/matplotlib-go/test/parity/mplot3d_plot3d"
	example_mplot3d_quiver3d "github.com/cwbudde/matplotlib-go/test/parity/mplot3d_quiver3d"
	example_mplot3d_scatter3d "github.com/cwbudde/matplotlib-go/test/parity/mplot3d_scatter3d"
	example_mplot3d_stem3d "github.com/cwbudde/matplotlib-go/test/parity/mplot3d_stem3d"
	example_mplot3d_surface3d "github.com/cwbudde/matplotlib-go/test/parity/mplot3d_surface3d"
	example_mplot3d_terrain "github.com/cwbudde/matplotlib-go/test/parity/mplot3d_terrain"
	example_mplot3d_text3d "github.com/cwbudde/matplotlib-go/test/parity/mplot3d_text3d"
	example_mplot3d_tricontour3d "github.com/cwbudde/matplotlib-go/test/parity/mplot3d_tricontour3d"
	example_mplot3d_tricontourf3d "github.com/cwbudde/matplotlib-go/test/parity/mplot3d_tricontourf3d"
	example_mplot3d_trisurf3d "github.com/cwbudde/matplotlib-go/test/parity/mplot3d_trisurf3d"
	example_mplot3d_voxels "github.com/cwbudde/matplotlib-go/test/parity/mplot3d_voxels"
	example_mplot3d_wire3d "github.com/cwbudde/matplotlib-go/test/parity/mplot3d_wire3d"
	example_multi_series_basic "github.com/cwbudde/matplotlib-go/test/parity/multi_series_basic"
	example_multi_series_color_cycle "github.com/cwbudde/matplotlib-go/test/parity/multi_series_color_cycle"
	example_named_colors "github.com/cwbudde/matplotlib-go/test/parity/named_colors"
	example_named_colors_gallery "github.com/cwbudde/matplotlib-go/test/parity/named_colors_gallery"
	example_patch_showcase "github.com/cwbudde/matplotlib-go/test/parity/patch_showcase"
	example_patch_style_matrix "github.com/cwbudde/matplotlib-go/test/parity/patch_style_matrix"
	example_path_clipped_transformed "github.com/cwbudde/matplotlib-go/test/parity/path_clipped_transformed"
	example_path_effects "github.com/cwbudde/matplotlib-go/test/parity/path_effects"
	example_pattern_gradient_effects "github.com/cwbudde/matplotlib-go/test/parity/pattern_gradient_effects"
	example_pcolor_flat "github.com/cwbudde/matplotlib-go/test/parity/pcolor_flat"
	example_pcolormesh_gouraud "github.com/cwbudde/matplotlib-go/test/parity/pcolormesh_gouraud"
	example_pcolormesh_masked "github.com/cwbudde/matplotlib-go/test/parity/pcolormesh_masked"
	example_pcolormesh_nearest "github.com/cwbudde/matplotlib-go/test/parity/pcolormesh_nearest"
	example_plot_variants "github.com/cwbudde/matplotlib-go/test/parity/plot_variants"
	example_polar_axes "github.com/cwbudde/matplotlib-go/test/parity/polar_axes"
	example_quad_mesh "github.com/cwbudde/matplotlib-go/test/parity/quad_mesh"
	example_radar_basic "github.com/cwbudde/matplotlib-go/test/parity/radar_basic"
	example_scale_asinh_ticks "github.com/cwbudde/matplotlib-go/test/parity/scale_asinh_ticks"
	example_scale_function_defaults "github.com/cwbudde/matplotlib-go/test/parity/scale_function_defaults"
	example_scale_logit_ticks "github.com/cwbudde/matplotlib-go/test/parity/scale_logit_ticks"
	example_scale_symlog_ticks "github.com/cwbudde/matplotlib-go/test/parity/scale_symlog_ticks"
	example_scatter_advanced "github.com/cwbudde/matplotlib-go/test/parity/scatter_advanced"
	example_scatter_basic "github.com/cwbudde/matplotlib-go/test/parity/scatter_basic"
	example_scatter_gallery "github.com/cwbudde/matplotlib-go/test/parity/scatter_gallery"
	example_scatter_marker_types "github.com/cwbudde/matplotlib-go/test/parity/scatter_marker_types"
	example_skewt_basic "github.com/cwbudde/matplotlib-go/test/parity/skewt_basic"
	example_specialty_artists "github.com/cwbudde/matplotlib-go/test/parity/specialty_artists"
	example_specialty_depth "github.com/cwbudde/matplotlib-go/test/parity/specialty_depth"
	example_spectrum_variants "github.com/cwbudde/matplotlib-go/test/parity/spectrum_variants"
	example_spy_image "github.com/cwbudde/matplotlib-go/test/parity/spy_image"
	example_spy_marker "github.com/cwbudde/matplotlib-go/test/parity/spy_marker"
	example_stat_variants "github.com/cwbudde/matplotlib-go/test/parity/stat_variants"
	example_stem_plot "github.com/cwbudde/matplotlib-go/test/parity/stem_plot"
	example_text_annotation_matrix "github.com/cwbudde/matplotlib-go/test/parity/text_annotation_matrix"
	example_text_labels_strict "github.com/cwbudde/matplotlib-go/test/parity/text_labels_strict"
	example_text_layout_gallery "github.com/cwbudde/matplotlib-go/test/parity/text_layout_gallery"
	example_ticks_styling_surface "github.com/cwbudde/matplotlib-go/test/parity/ticks_styling_surface"
	example_title_strict "github.com/cwbudde/matplotlib-go/test/parity/title_strict"
	example_transform_annotation_modes "github.com/cwbudde/matplotlib-go/test/parity/transform_annotation_modes"
	example_transform_coordinates "github.com/cwbudde/matplotlib-go/test/parity/transform_coordinates"
	example_twoslope_norm_image "github.com/cwbudde/matplotlib-go/test/parity/twoslope_norm_image"
	example_units_categories "github.com/cwbudde/matplotlib-go/test/parity/units_categories"
	example_units_custom_converter "github.com/cwbudde/matplotlib-go/test/parity/units_custom_converter"
	example_units_dates "github.com/cwbudde/matplotlib-go/test/parity/units_dates"
	example_units_overview "github.com/cwbudde/matplotlib-go/test/parity/units_overview"
	example_unstructured_showcase "github.com/cwbudde/matplotlib-go/test/parity/unstructured_showcase"
	example_vector_fields "github.com/cwbudde/matplotlib-go/test/parity/vector_fields"
	example_widgets_gallery "github.com/cwbudde/matplotlib-go/test/parity/widgets_gallery"
)

// Case describes a runnable Go/Python parity example.
type Case struct {
	ID               string
	Title            string
	Topic            string
	GoSourcePath     string
	PythonSourcePath string
}

var renderByID = map[string]func() image.Image{
	"animation_gallery":                   showcase_animation_gallery.Render,
	"basic_line":                          example_basic_line.Render,
	"basic_line_labels":                   example_basic_line_labels.Render,
	"joins_caps":                          example_joins_caps.Render,
	"dashes":                              example_dashes.Render,
	"line2d_semantics":                    example_line2d_semantics.Render,
	"line2d_markers":                      example_line2d_markers.Render,
	"path_effects":                        example_path_effects.Render,
	"pattern_gradient_effects":            example_pattern_gradient_effects.Render,
	"scatter_basic":                       example_scatter_basic.Render,
	"scatter_marker_types":                example_scatter_marker_types.Render,
	"scatter_advanced":                    example_scatter_advanced.Render,
	"scatter_gallery":                     example_scatter_gallery.Render,
	"bar_basic_frame":                     example_bar_basic_frame.Render,
	"bar_basic_ticks":                     example_bar_basic_ticks.Render,
	"bar_basic_tick_labels":               example_bar_basic_tick_labels.Render,
	"bar_basic_title":                     example_bar_basic_title.Render,
	"bar_basic":                           example_bar_basic.Render,
	"bar_horizontal":                      example_bar_horizontal.Render,
	"bar_grouped":                         example_bar_grouped.Render,
	"fill_basic":                          example_fill_basic.Render,
	"fill_between":                        example_fill_between.Render,
	"fill_stacked":                        example_fill_stacked.Render,
	"errorbar_basic":                      example_errorbar_basic.Render,
	"lines_markers_gallery":               example_lines_markers_gallery.Render,
	"multi_series_basic":                  example_multi_series_basic.Render,
	"multi_series_color_cycle":            example_multi_series_color_cycle.Render,
	"legend_layout_matrix":                example_legend_layout_matrix.Render,
	"text_annotation_matrix":              example_text_annotation_matrix.Render,
	"hist_basic":                          example_hist_basic.Render,
	"hist_density":                        example_hist_density.Render,
	"hist_strategies":                     example_hist_strategies.Render,
	"boxplot_basic":                       example_boxplot_basic.Render,
	"text_labels_strict":                  example_text_labels_strict.Render,
	"title_strict":                        example_title_strict.Render,
	"image_heatmap":                       example_image_heatmap.Render,
	"image_variants_gallery":              example_image_variants_gallery.Render,
	"imshow_clipped":                      example_imshow_clipped.Render,
	"imshow_transformed":                  example_imshow_transformed.Render,
	"imshow_bilinear":                     example_imshow_bilinear.Render,
	"imshow_bicubic":                      example_imshow_bicubic.Render,
	"imshow_interpolation_matrix":         example_imshow_interpolation_matrix.Render,
	"image_alpha":                         example_image_alpha.Render,
	"matshow_basic":                       example_matshow_basic.Render,
	"spy_marker":                          example_spy_marker.Render,
	"spy_image":                           example_spy_image.Render,
	"colormap_diverging":                  example_colormap_diverging.Render,
	"colormap_qualitative":                example_colormap_qualitative.Render,
	"colormap_cyclic":                     example_colormap_cyclic.Render,
	"colormap_families_gallery":           example_colormap_families_gallery.Render,
	"named_colors":                        example_named_colors.Render,
	"named_colors_gallery":                example_named_colors_gallery.Render,
	"mathtext_basic":                      example_mathtext_basic.Render,
	"mathtext_fractions":                  example_mathtext_fractions.Render,
	"mathtext_integrals":                  example_mathtext_integrals.Render,
	"mathtext_matrices":                   example_mathtext_matrices.Render,
	"mathtext_inline_labels":              example_mathtext_inline_labels.Render,
	"mathtext_gallery":                    example_mathtext_gallery.Render,
	"text_layout_gallery":                 example_text_layout_gallery.Render,
	"axes_top_right_inverted":             example_axes_top_right_inverted.Render,
	"axes_control_surface":                example_axes_control_surface.Render,
	"axes_convenience_helpers":            example_axes_convenience_helpers.Render,
	"axes_option_breadth_17_75_3":         example_axes_option_breadth_17_75_3.Render,
	"transform_coordinates":               example_transform_coordinates.Render,
	"transform_annotation_modes":          example_transform_annotation_modes.Render,
	"path_clipped_transformed":            example_path_clipped_transformed.Render,
	"layout_bbox_helpers":                 example_layout_bbox_helpers.Render,
	"formatter_engineering_labels":        example_formatter_engineering_labels.Render,
	"formatter_fixed_null_labels":         example_formatter_fixed_null_labels.Render,
	"formatter_log_mathtext_labels":       example_formatter_log_mathtext_labels.Render,
	"formatter_percent_labels":            example_formatter_percent_labels.Render,
	"formatter_scalar_scientific_labels":  example_formatter_scalar_scientific_labels.Render,
	"locator_fixed_index_labels":          example_locator_fixed_index_labels.Render,
	"locator_linear_labels":               example_locator_linear_labels.Render,
	"locator_log_minor_threshold_labels":  example_locator_log_minor_threshold_labels.Render,
	"locator_maxn_edge_labels":            example_locator_maxn_edge_labels.Render,
	"scale_asinh_ticks":                   example_scale_asinh_ticks.Render,
	"scale_function_defaults":             example_scale_function_defaults.Render,
	"scale_logit_ticks":                   example_scale_logit_ticks.Render,
	"scale_symlog_ticks":                  example_scale_symlog_ticks.Render,
	"ticks_styling_surface":               example_ticks_styling_surface.Render,
	"artist_metadata":                     example_artist_metadata.Render,
	"gridspec_composition":                example_gridspec_composition.Render,
	"figure_labels_composition":           example_figure_labels_composition.Render,
	"colorbar_composition":                example_colorbar_composition.Render,
	"colorbar_variants_gallery":           example_colorbar_variants_gallery.Render,
	"annotation_composition":              example_annotation_composition.Render,
	"annotation_legend_offsetbox_gallery": example_annotation_legend_offsetbox_gallery.Render,
	"patch_showcase":                      example_patch_showcase.Render,
	"patch_style_matrix":                  example_patch_style_matrix.Render,
	"mesh_contour_tri":                    example_mesh_contour_tri.Render,
	"plot_variants":                       example_plot_variants.Render,
	"spectrum_variants":                   example_spectrum_variants.Render,
	"stat_variants":                       example_stat_variants.Render,
	"specialty_depth":                     example_specialty_depth.Render,
	"stem_plot":                           example_stem_plot.Render,
	"specialty_artists":                   example_specialty_artists.Render,
	"date_concise_intraday_labels":        example_date_concise_intraday_labels.Render,
	"date_month_year_labels":              example_date_month_year_labels.Render,
	"units_overview":                      example_units_overview.Render,
	"units_dates":                         example_units_dates.Render,
	"units_categories":                    example_units_categories.Render,
	"units_custom_converter":              example_units_custom_converter.Render,
	"vector_fields":                       example_vector_fields.Render,
	"polar_axes":                          example_polar_axes.Render,
	"geo_mollweide_axes":                  example_geo_mollweide_axes.Render,
	"geo_aitoff_axes":                     example_geo_aitoff_axes.Render,
	"geo_hammer_axes":                     example_geo_hammer_axes.Render,
	"geo_lambert_axes":                    example_geo_lambert_axes.Render,
	"radar_basic":                         example_radar_basic.Render,
	"skewt_basic":                         example_skewt_basic.Render,
	"mplot3d_basic":                       example_mplot3d_basic.Render,
	"mplot3d_terrain":                     example_mplot3d_terrain.Render,
	"mplot3d_gallery":                     example_mplot3d_gallery.Render,
	"mplot3d_plot3d":                      example_mplot3d_plot3d.Render,
	"mplot3d_scatter3d":                   example_mplot3d_scatter3d.Render,
	"mplot3d_surface3d":                   example_mplot3d_surface3d.Render,
	"mplot3d_wire3d":                      example_mplot3d_wire3d.Render,
	"mplot3d_trisurf3d":                   example_mplot3d_trisurf3d.Render,
	"mplot3d_tricontour3d":                example_mplot3d_tricontour3d.Render,
	"mplot3d_tricontourf3d":               example_mplot3d_tricontourf3d.Render,
	"mplot3d_text3d":                      example_mplot3d_text3d.Render,
	"mplot3d_bar3d":                       example_mplot3d_bar3d.Render,
	"mplot3d_bar2d_zdir":                  example_mplot3d_bar2d_zdir.Render,
	"mplot3d_voxels":                      example_mplot3d_voxels.Render,
	"mplot3d_quiver3d":                    example_mplot3d_quiver3d.Render,
	"mplot3d_contour3d":                   example_mplot3d_contour3d.Render,
	"mplot3d_contourf3d":                  example_mplot3d_contourf3d.Render,
	"mplot3d_errorbar3d":                  example_mplot3d_errorbar3d.Render,
	"mplot3d_stem3d":                      example_mplot3d_stem3d.Render,
	"mplot3d_fill_between3d":              example_mplot3d_fill_between3d.Render,
	"unstructured_showcase":               example_unstructured_showcase.Render,
	"arrays_showcase":                     example_arrays_showcase.Render,
	"widgets_gallery":                     example_widgets_gallery.Render,
	"axisartist_showcase":                 example_axisartist_showcase.Render,
	"axes_grid1_showcase":                 example_axes_grid1_showcase.Render,
	"pcolor_flat":                         example_pcolor_flat.Render,
	"pcolormesh_nearest":                  example_pcolormesh_nearest.Render,
	"pcolormesh_gouraud":                  example_pcolormesh_gouraud.Render,
	"pcolormesh_masked":                   example_pcolormesh_masked.Render,
	"hist2d_weighted_density":             example_hist2d_weighted_density.Render,
	"asinh_norm_image":                    example_asinh_norm_image.Render,
	"boundarynorm_pcolormesh":             example_boundarynorm_pcolormesh.Render,
	"collection_mutable_scalarmap":        example_collection_mutable_scalarmap.Render,
	"colorbar_boundary_values":            example_colorbar_boundary_values.Render,
	"colorbar_horizontal_ticks":           example_colorbar_horizontal_ticks.Render,
	"lognorm_imshow":                      example_lognorm_imshow.Render,
	"twoslope_norm_image":                 example_twoslope_norm_image.Render,
	"colorbar_extensions":                 example_colorbar_extensions.Render,
	"large_scatter":                       example_large_scatter.Render,
	"mixed_collection":                    example_mixed_collection.Render,
	"mixed_raster_vector":                 example_mixed_raster_vector.Render,
	"quad_mesh":                           example_quad_mesh.Render,
	"gouraud_triangles":                   example_gouraud_triangles.Render,
	"clip_path_batch":                     example_clip_path_batch.Render,
}

var figureByID = map[string]func() *core.Figure{
	"animation_gallery":                   showcase_animation_gallery.Plot,
	"annotation_composition":              showcase_annotation_composition.Plot,
	"artist_metadata":                     example_artist_metadata.Plot,
	"line2d_markers":                      example_line2d_markers.Plot,
	"line2d_semantics":                    example_line2d_semantics.Plot,
	"path_effects":                        example_path_effects.Plot,
	"pattern_gradient_effects":            example_pattern_gradient_effects.Plot,
	"arrays_showcase":                     showcase_arrays_showcase.Plot,
	"axes_option_breadth_17_75_3":         example_axes_option_breadth_17_75_3.Plot,
	"axes_grid1_showcase":                 showcase_axes_grid1_showcase.Plot,
	"axes_control_surface":                showcase_axes_control_surface.Plot,
	"transform_annotation_modes":          example_transform_annotation_modes.Plot,
	"path_clipped_transformed":            example_path_clipped_transformed.Plot,
	"layout_bbox_helpers":                 example_layout_bbox_helpers.Plot,
	"formatter_engineering_labels":        example_formatter_engineering_labels.Plot,
	"formatter_fixed_null_labels":         example_formatter_fixed_null_labels.Plot,
	"formatter_log_mathtext_labels":       example_formatter_log_mathtext_labels.Plot,
	"formatter_percent_labels":            example_formatter_percent_labels.Plot,
	"formatter_scalar_scientific_labels":  example_formatter_scalar_scientific_labels.Plot,
	"locator_fixed_index_labels":          example_locator_fixed_index_labels.Plot,
	"locator_linear_labels":               example_locator_linear_labels.Plot,
	"locator_log_minor_threshold_labels":  example_locator_log_minor_threshold_labels.Plot,
	"locator_maxn_edge_labels":            example_locator_maxn_edge_labels.Plot,
	"scale_asinh_ticks":                   example_scale_asinh_ticks.Plot,
	"scale_function_defaults":             example_scale_function_defaults.Plot,
	"scale_logit_ticks":                   example_scale_logit_ticks.Plot,
	"scale_symlog_ticks":                  example_scale_symlog_ticks.Plot,
	"ticks_styling_surface":               example_ticks_styling_surface.Plot,
	"axisartist_showcase":                 showcase_axisartist_showcase.Plot,
	"asinh_norm_image":                    example_asinh_norm_image.Plot,
	"basic_line":                          showcase_basic_line.Plot,
	"basic_line_labels":                   showcase_basic_line_labels.Plot,
	"bar_basic":                           showcase_bar_basic.Plot,
	"boxplot_basic":                       showcase_boxplot_basic.Plot,
	"colorbar_composition":                showcase_colorbar_composition.Plot,
	"colorbar_variants_gallery":           showcase_colorbar_variants_gallery.Plot,
	"colorbar_boundary_values":            example_colorbar_boundary_values.Plot,
	"colorbar_horizontal_ticks":           example_colorbar_horizontal_ticks.Plot,
	"colormap_diverging":                  example_colormap_diverging.Plot,
	"colormap_families_gallery":           showcase_colormap_families_gallery.Plot,
	"collection_mutable_scalarmap":        example_collection_mutable_scalarmap.Plot,
	"errorbar_basic":                      showcase_errorbar_basic.Plot,
	"fill_between":                        example_fill_between.Plot,
	"geo_mollweide_axes":                  showcase_geo_mollweide_axes.Plot,
	"gridspec_composition":                showcase_gridspec_composition.Plot,
	"scatter_basic":                       showcase_scatter_basic.Plot,
	"scatter_gallery":                     showcase_scatter_gallery.Plot,
	"lines_markers_gallery":               showcase_lines_markers_gallery.Plot,
	"hist_basic":                          showcase_hist_basic.Plot,
	"mesh_contour_tri":                    showcase_mesh_contour_tri.Plot,
	"image_heatmap":                       showcase_image_heatmap.Plot,
	"image_variants_gallery":              showcase_image_variants_gallery.Plot,
	"mathtext_gallery":                    showcase_mathtext_gallery.Plot,
	"text_layout_gallery":                 showcase_text_layout_gallery.Plot,
	"multi_series_basic":                  showcase_multi_series_basic.Plot,
	"legend_layout_matrix":                example_legend_layout_matrix.Plot,
	"text_annotation_matrix":              example_text_annotation_matrix.Plot,
	"named_colors":                        example_named_colors.Plot,
	"named_colors_gallery":                showcase_named_colors_gallery.Plot,
	"annotation_legend_offsetbox_gallery": showcase_annotation_legend_offsetbox_gallery.Plot,
	"plot_variants":                       showcase_plot_variants.Plot,
	"polar_axes":                          showcase_polar_axes.Plot,
	"radar_basic":                         showcase_radar_basic.Plot,
	"patch_showcase":                      example_patch_showcase.Plot,
	"patch_style_matrix":                  example_patch_style_matrix.Plot,
	"skewt_basic":                         showcase_skewt_basic.Plot,
	"specialty_artists":                   showcase_specialty_artists.Plot,
	"stat_variants":                       showcase_stat_variants.Plot,
	"text_labels_strict":                  example_text_labels_strict.Plot,
	"date_concise_intraday_labels":        example_date_concise_intraday_labels.Plot,
	"date_month_year_labels":              example_date_month_year_labels.Plot,
	"units_overview":                      showcase_units_overview.Plot,
	"vector_fields":                       showcase_vector_fields.Plot,
	"mathtext_basic":                      example_mathtext_basic.Plot,
	"mathtext_fractions":                  example_mathtext_fractions.Plot,
	"mathtext_integrals":                  example_mathtext_integrals.Plot,
	"mathtext_matrices":                   example_mathtext_matrices.Plot,
	"mathtext_inline_labels":              example_mathtext_inline_labels.Plot,
	"mplot3d_basic":                       example_mplot3d_basic.Plot,
	"mixed_collection":                    example_mixed_collection.Plot,
	"mixed_raster_vector":                 showcase_mixed_raster_vector.Plot,
	"mplot3d_gallery":                     showcase_mplot3d_gallery.Plot,
	"spectrum_variants":                   example_spectrum_variants.Plot,
	"unstructured_showcase":               showcase_unstructured_showcase.Plot,
	"widgets_gallery":                     example_widgets_gallery.Plot,
	"imshow_clipped":                      example_imshow_clipped.Plot,
	"imshow_transformed":                  example_imshow_transformed.Plot,
	"imshow_interpolation_matrix":         example_imshow_interpolation_matrix.Plot,
}

// Cases returns the canonical parity examples in catalog order.
func Cases() []Case {
	catalog := examplecatalog.Cases()
	out := make([]Case, 0, len(catalog))
	for _, c := range catalog {
		out = append(out, Case{
			ID:               c.ID,
			Title:            c.Title,
			Topic:            c.Topic,
			GoSourcePath:     GoSourcePath(c.ID),
			PythonSourcePath: PythonSourcePath(c.ID),
		})
	}
	return out
}

// Lookup finds a parity example by case ID.
func Lookup(id string) (Case, bool) {
	for _, c := range Cases() {
		if c.ID == id {
			return c, true
		}
	}
	return Case{}, false
}

// Render renders a parity example by case ID.
func Render(id string) (image.Image, Case, error) {
	c, ok := Lookup(id)
	if !ok {
		return nil, Case{}, fmt.Errorf("parity: unknown case %q", id)
	}
	render, ok := renderByID[id]
	if !ok {
		return nil, Case{}, fmt.Errorf("parity: missing Go renderer for case %q", id)
	}
	return render(), c, nil
}

// Figure builds a backend-agnostic figure for parity cases that expose one.
func Figure(id string) (*core.Figure, Case, error) {
	c, ok := Lookup(id)
	if !ok {
		return nil, Case{}, fmt.Errorf("parity: unknown case %q", id)
	}
	figure, ok := figureByID[id]
	if !ok {
		return nil, Case{}, fmt.Errorf("parity: missing Go figure factory for case %q", id)
	}
	return figure(), c, nil
}

// RenderToFile renders a parity example to outputDir/<id>.png.
func RenderToFile(id, outputDir string) (string, error) {
	img, c, err := Render(id)
	if err != nil {
		return "", err
	}
	if outputDir == "" {
		outputDir = "."
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("create output dir %s: %w", outputDir, err)
	}
	path := filepath.Join(outputDir, c.ID+".png")
	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		return "", fmt.Errorf("encode %s: %w", path, err)
	}
	return path, nil
}

// GoSourcePath returns the repository-relative canonical Go source path.
func GoSourcePath(id string) string {
	return "test/parity/" + id + "/plot.go"
}

// PythonSourcePath returns the repository-relative canonical Python source path.
func PythonSourcePath(id string) string {
	return "test/parity/" + id + "/plot.py"
}
