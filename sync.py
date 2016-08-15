# -*- encoding:utf-8 -*-

import os, uuid, json, shutil
import re
import requests
from requests_toolbelt.multipart.encoder import MultipartEncoder

class SyncMgr(object):
	def __init__(self, remoteAddress, localStorage):
		self.remoteAddress = remoteAddress
		self.localStorage = localStorage

	def _listLocal(self):
		pathMap = {}
		blogAddress = os.path.join(self.localStorage, "blog")
		for lists in os.listdir(blogAddress): 
			path = os.path.join(blogAddress, lists)
			if os.path.isdir(path): 
				pathMap[path] = []
				for blog in os.listdir(path):
					blogPath = os.path.join(path, blog)
					if os.path.isdir(blogPath):
						print blogPath
						pathMap[path].append({"name": blog, "path": blogPath, "sort": lists})
		return pathMap

	def _preHandleHtml(self, path):
		f = open(path)
		content = f.read()
		f.close()
		allImage = []
		p = re.compile(r"<img.*?/>")
		p1 = re.compile(r"src='(.*?)'")
		for s in p.findall(content):
			id = str(uuid.uuid1())
			newImagePath = "./img?id=%s" % id
			allImage.append({"id": id, "path": p1.findall(s)[0]})
			content = content.replace(s, '<img src="%s" />' % newImagePath)
		f = open(path, "w")
		f.write(content)
		f.close()
		print open(path).read()
		return allImage

	def uploadBlog(self):
		filePath = self._listLocal()
		print 'upload blog, please input index.'
		index = 0
		blogList = []
		for (sort, blogs) in filePath.items():
			print sort
			for blog in blogs:
				index = index + 1
				blogList.append(blog)
				print "%d. %s" % (index, blog["name"])
		input_index = raw_input("index: ")
		try:
			input_index = int(input_index) - 1
		except Exception, e:
			print 'Input error'
			return

		rawBlogFolder = blogList[input_index]["path"]
		rawBlogName = blogList[input_index]["name"]
		print type(rawBlogFolder)
		blogInfoPath = os.path.join(rawBlogFolder, "blog.info")
		# read uuid
		blogInfoContent = open(blogInfoPath)
		js = json.loads(blogInfoContent.read())
		if js["uuid"] == "":
			id = str(uuid.uuid1())
			js["uuid"] = id
			blogInfoContent.close()
			blogInfoContent = open(blogInfoPath, "w+")
			blogInfoContent.write(json.dumps(js))
		else:
			id = js["uuid"]
		blogInfoContent.close()

		# copy to tmp directory
		tmpFilePath = "/tmp/%s" %  id
		print rawBlogFolder
		print tmpFilePath
		if os.path.exists(tmpFilePath):
			shutil.rmtree(tmpFilePath)
		shutil.copytree(rawBlogFolder.decode("utf-8"), tmpFilePath)

		htmlPath = ''
		for file in os.listdir(tmpFilePath):
			path = os.path.join(tmpFilePath, file)
			if os.path.isfile(path):
				if file == ".DS_Store":
					os.remove(path)
				elif file == "blog.info" or file == "cover.img":
					pass
				else:
					fname, ext = os.path.splitext(file)
					newPath = os.path.join(tmpFilePath, id + ext)
					if ext == ".html":
						htmlPath = newPath
					os.rename(path, newPath)

		os.chdir("/tmp")

		# compress img
		tmpImagePath = "img"
		tmpImageCompressPath = "img.tar.gz"
		if os.path.exists(tmpImagePath):
			shutil.rmtree(tmpImagePath)
		os.mkdir(tmpImagePath)
		allImage = self._preHandleHtml(htmlPath)

		coverImgPath = os.path.join(tmpFilePath, "cover.img")
		if os.path.exists(coverImgPath):
			allImage.append({"id": id, "path": coverImgPath})

		for image in allImage:
			tmpPath = os.path.join(tmpImagePath, image["id"])
			shutil.copyfile(image["path"], tmpPath)
		command = "tar -zcvf %s %s" % (tmpImageCompressPath, tmpImagePath)
		os.system(command)
		imgContent = open(tmpImageCompressPath)

		# compress blog
		tmpBlogCompressPath = "%s.tar.gz" % id
		sort = blogList[input_index]["sort"]
		command = "tar -zcvf %s %s" % (tmpBlogCompressPath, id)
		os.system(command)
		fileContent = open(tmpBlogCompressPath)

		# send request
		multipart_data = MultipartEncoder(fields = {
			# 'title': rawBlogName, 
			# 'sort': sort,
			'uuid': js["uuid"],
			# 'tag': js["tag"],
			# 'file': fileContent,
			# 'img': imgContent,
		})

		response = requests.post(self.remoteAddress, data = multipart_data,
                  headers={'Content-Type': multipart_data.content_type})

		print response

		shutil.rmtree(tmpImagePath)
		os.remove(tmpImageCompressPath)
		os.remove(tmpBlogCompressPath)

def main():
	mgr = SyncMgr("http://112.124.24.31:80/personal/file-upload", "/Users/sjjwind/blog-storage")
	# mgr = SyncMgr("http://www.baidu.com", "/Users/sjjwind/storage.blog")
	mgr.uploadBlog()

if __name__ == '__main__':
	main()
