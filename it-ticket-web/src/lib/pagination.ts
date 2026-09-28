export const PAGE_SIZE_OPTIONS = [10, 20, 50]

export function listPagination(
  page: number,
  pageSize: number,
  total: number,
  onChange: (nextPage: number, nextSize: number) => void,
) {
  return {
    current: page,
    pageSize,
    total,
    showSizeChanger: true,
    pageSizeOptions: PAGE_SIZE_OPTIONS,
    showTotal: (count: number) => `共 ${count} 条`,
    onChange: (nextPage: number, nextSize: number) => {
      onChange(nextSize === pageSize ? nextPage : 1, nextSize)
    },
  }
}
